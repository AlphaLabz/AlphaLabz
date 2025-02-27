package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleUserRemove(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from header and validate it
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hasPermmission, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "delete",
		Scopes:    "*",
	})
	if err != nil || !hasPermmission {
		fmt.Println(hasPermmission)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get delete user Id from request url
	userId := r.URL.Query().Get("id")
	if userId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// list all users before deleting the user to ensure it exists
	users, totalCount, err := pbClient.ListUsers([]string{"id", "role", "name"}, nil, fmt.Sprintf("(id='%s')", userId))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	// Check if the user exists. If not, return error.
	if totalCount != 1 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is an admin before deleting it. If yes, return error.
	if users[0].RoleId == "0001" {
		http.Error(w, "Cannot delete admin user", http.StatusForbidden)
		return
	}

	if err := pbClient.DeleteUser(userId); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	// Response
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}
