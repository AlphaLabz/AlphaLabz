package role

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
)

// Delete a Role
// Only users with the delete:"custom" permission on the "roles" resource can delete a role.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `DELETE`
//
// ✅ URL Parameter:
//   - `id` (string, required) → The ID of the role to be deleted.
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "message": "Role deleted successfully"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Invalid role ID or attempt to delete a system role.
//   - 401 Unauthorized → Missing or Invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 404 Not Found → Role does not exist.
//   - 405 Method Not Allowed → Invalid HTTP method (only DELETE is allowed).
//   - 500 Internal Server Error → Server issue or failure deleting the role.
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
