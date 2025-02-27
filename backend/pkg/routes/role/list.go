package role

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"net/http"
)

type rolesResp struct {
	Roles []pocketbase.Role
}

func HandleRoleList(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	var permissionConfig = casbin.PermissionConfig{
		Resources: "roles",
		Actions:   "list",
	}

	// Check if the request method is GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from request header
	rawJwtToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userRole, err := pbClient.FetchUserRole(rawJwtToken)
	if err != nil {
		http.Error(w, "Failed to fetch user role", http.StatusInternalServerError)
		return
	}

	scopes, err := ce.CheckPermissionScopes(userRole.RoleId, permissionConfig.Resources, permissionConfig.Actions)
	if err != nil {
		http.Error(w, "Failed to check permission", http.StatusInternalServerError)
	} else {
		if len(scopes) == 0 {
			http.Error(w, "No permission to access this resource", http.StatusForbidden)
			return
		}
	}

	roles, err := pbClient.GetAvailableRoles(scopes)
	if err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
	}

	result := rolesResp{
		Roles: roles,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}
