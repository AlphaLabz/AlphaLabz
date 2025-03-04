package labbook

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"net/http"
)

func HandleLabBookView(w http.ResponseWriter, r *http.Request, labbookId string, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
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

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Failed to obtain userId from token", http.StatusInternalServerError)
	}

	hasStarPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "view",
		Scopes:    "*",
	})
	if err != nil {
		http.Error(w, "Failed to verify permission", http.StatusInternalServerError)
	}

	labbookContent, err := pbClient.ViewLabbook(labbookId, []string{"*"})
	if err != nil {
		http.Error(w, "Failed to view labbook", http.StatusInternalServerError)
	}

	// if userId in access list or hasStarPermission
	if hasStarPermission || tools.Contains(labbookContent.ShareWith, userId) {
		json.NewEncoder(w).Encode(labbookContent)
	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}
