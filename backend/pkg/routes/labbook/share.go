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

// Share Lab Book
// Only users with the update:"share" permission on the "lab_books" resource can share a lab book with another user.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `POST`
//
// ✅ Request Body: `Content-Type: application/json`
// - Fields:
//   - `LabbookId` (string, required) → The ID of the lab book to be shared.
//   - `RecipientId` (string, required) → The ID of the user receiving access to the lab book.
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "message": "Labbook shared successfully"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing required fields or invalid request body format.
//   - 401 Unauthorized → Missing or Invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only POST is allowed).
//   - 404 Not Found → Recipient does not exist.
//   - 409 Conflict → Recipient already has access to the lab book.
//   - 500 Internal Server Error → Server issue or database operation failure.
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

// Get Shared Lab Books List
// Only users with the view:"shared" permission on the "lab_books" resource can retrieve the list of lab books that have been shared with them.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `GET`
//
// ✅ Successful Response (200 OK):
// Returns a JSON array containing the lab books that have been shared with the authenticated user.
//
// ❌ Error Responses:
//   - 401 Unauthorized → Missing or invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only GET is allowed).
//   - 500 Internal Server Error → Server issue or failure in retrieving shared lab books.
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
