package role

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleDeleteRole(w http.ResponseWriter, r *http.Request, id string, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if id == "" {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "roles",
		Actions:   "delete",
		Scopes:    "custom",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	roles, err := pbClient.ListRoles([]string{"id", "type"}, fmt.Sprintf("id='%s'", id))
	if err != nil {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}

	if len(roles) == 0 {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}

	if roles[0].Type != "custom" {
		http.Error(w, "Cannot delete system role", http.StatusBadRequest)
		return
	}

	if err := pbClient.DeleteRole(id); err != nil {
		http.Error(w, "Failed to delete role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Role deleted successfully"})
}
