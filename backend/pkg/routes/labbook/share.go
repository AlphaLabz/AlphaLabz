package labbook

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
)

type ShareRequest struct {
	LabbookId   string `json:"labbook_id"`
	RecipientId string `json:"recipient_id"`
}

func HandleShareLabbook(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
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
	hasPermission, starPermission, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "update",
		Scopes:    "share",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body into a struct
	var shareRequest ShareRequest
	err = json.NewDecoder(r.Body).Decode(&shareRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if shareRequest.LabbookId == "" || shareRequest.RecipientId == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	labbookInfo, err := pbClient.ViewLabbook(shareRequest.LabbookId, []string{"id", "creator", "share_to"})
	if err != nil {
		http.Error(w, "Failed to retrieve labbook information", http.StatusInternalServerError)
		return
	}

	if tools.Contains(labbookInfo.ShareWith, shareRequest.RecipientId) || labbookInfo.Creator == shareRequest.RecipientId || labbookInfo.Reviewer == shareRequest.RecipientId {
		http.Error(w, "Recipient already has access", http.StatusConflict)
		return
	}

	requester, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Failed to obtain requester ID from token", http.StatusInternalServerError)
		return
	}

	if labbookInfo.Creator != requester && !starPermission {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipientExist, err := pbClient.CheckUserExists(shareRequest.RecipientId)
	if err != nil {
		http.Error(w, "Failed to check if recipient exists", http.StatusInternalServerError)
		return
	}

	if !recipientExist {
		http.Error(w, "Recipient does not exist", http.StatusNotFound)
		return
	} else {
		err := pbClient.ShareLabbook(shareRequest.LabbookId, shareRequest.RecipientId, labbookInfo.ShareWith)
		if err != nil {
			http.Error(w, "Failed to share labbook", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Labbook shared successfully"})
}

func GetSharedList(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
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
		Actions:   "view",
		Scopes:    "shared",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get the lab books that have been shared with the user from the database

	sharedLabbooks, err := pbClient.ListLabbooks(fmt.Sprintf("share_with?~'%s'", userId), []string{"*"})
	if err != nil {
		http.Error(w, "Failed to get shared labbooks", http.StatusInternalServerError)
		return
	}

	// Return the list of shared labbooks as JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sharedLabbooks)
}
