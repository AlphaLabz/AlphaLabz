package role

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateNewRole(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "roles",
		Actions:   "create",
		Scopes:    "custom",
	})
	if err != nil || !hasPermission {
		fmt.Print("Permission verification failed: ", err)
		fmt.Println(hasPermission)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var newRole pocketbase.NewRoleRequest
	err = json.NewDecoder(r.Body).Decode(&newRole)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = pbClient.CreateRole(newRole)
	if err != nil {
		roleInfo, err := pbClient.ListRoles([]string{"name"}, fmt.Sprintf("name=%s", newRole.Name))
		if err != nil || len(roleInfo) > 0 {
			http.Error(w, "Role already exists", http.StatusConflict)
			return
		} else {
			http.Error(w, "Failed to create role", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Role created successfully"))
}
