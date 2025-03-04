package role

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
)

// Create a New Role
// Only users with the create:"custom" permission on the "roles" resource can create a new role.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `POST`
//
// ✅ Request Body: `Content-Type: application/json`
// - Fields:
//   - `Name` (string, required) → The name of the new role.
//   - `Permissions` (array, required) → A list of permissions assigned to the new role.
//
// ✅ Successful Response (201 Created):
//
//	{
//	    "message": "Role created successfully"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing required fields or invalid request body format.
//   - 401 Unauthorized → Missing or Invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only POST is allowed).
//   - 409 Conflict → Role with the same name already exists.
//   - 500 Internal Server Error → Server issue or failure creating the role.
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
