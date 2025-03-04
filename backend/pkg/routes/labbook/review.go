package labbook

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type labbookReviewRequest struct {
	LabbookId string `json:"labbook_id"`
	Status    string `json:"status"`
	Comment   string `json:"comment"`
}

func HandleLabBookReview(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is PATCH
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get request authorization header
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Verify user permission using Casbin enforcer
	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "update",
		Scopes:    "status"})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Decode the JSON request body into a LabBookReviewRequest struct
	var reviewRequest labbookReviewRequest
	err = json.NewDecoder(r.Body).Decode(&reviewRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the labbook ID and status fields in the request
	if reviewRequest.LabbookId == "" || (reviewRequest.Status != "approved" && reviewRequest.Status != "Rejected") {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve lab book information from PocketBase database
	labbook, err := pbClient.ViewLabbook(reviewRequest.LabbookId, []string{"id", "reviewer", "status"})
	if err != nil {
		http.Error(w, "Failed to retrieve lab book", http.StatusInternalServerError)
		return
	}

	if labbook.ReviewStatus != "pending" {
		http.Error(w, "The lab book aleardy been verified", http.StatusConflict)
		return
	}

	// Extract userId from JWT token
	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Verify that the reviewer is the same as the one in the request body
	if labbook.Reviewer != userId {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update the lab book review status and comments
	if err = pbClient.UpdateLabbook(reviewRequest.LabbookId, map[string]interface{}{
		"review_status":  reviewRequest.Status,
		"review_comment": reviewRequest.Comment,
	}); err != nil {
		http.Error(w, "Failed to update lab book", http.StatusInternalServerError)
		return
	}

	// Encode response with success message
	json.NewEncoder(w).Encode(map[string]string{"message": "Lab book review updated successfully"})
}

func GetAvailiableReviewers(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is GET.
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get request authorization header
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Verify user permission using Casbin enforcer
	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "create",
		Scopes:    "own",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get users with permission to update lab books status
	hasPermissionList, err := ce.GetRoleIDsByPermission("lab_books", "update", "status")
	if err != nil {
		http.Error(w, "Failed to get role IDs by permission", http.StatusInternalServerError)
		return
	}

	// Format the role filter correctly using OR conditions
	roleConditions := []string{}
	for _, roleID := range hasPermissionList {
		roleConditions = append(roleConditions, fmt.Sprintf("role='%s'", roleID))
	}

	// Combine conditions with OR (||)
	rawFilter := "(" + strings.Join(roleConditions, " || ") + ")"

	// Properly encode without adding unwanted `+`
	roleFilter := strings.ReplaceAll(url.QueryEscape(rawFilter), "+", "%20")

	reviewers, _, err := pbClient.ListUsers([]string{"id", "name", "role"}, []string{}, roleFilter)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "success", "reviewers": reviewers})
}

func GetPendingReviews(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Failed to extract token", http.StatusInternalServerError)
		return
	}

	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "update",
		Scopes:    "review",
	})
	if !hasPermission || err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get the lab book upload history from the database
	filter := fmt.Sprintf("creator='%s' && review_status='pending'", userId)
	encodedFilter := url.QueryEscape(filter)

	pendingRievews, err := pbClient.ListLabbooks(encodedFilter, []string{"*"})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get lab book upload history", http.StatusInternalServerError)
		return
	}

	// Return the lab book upload history as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pendingRievews)
}
