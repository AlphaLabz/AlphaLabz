package role

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"net/http"
)

// List Roles
// Only users with the list permission on the "roles" resource can retrieve the list of roles they have access to.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `GET`
//
// ✅ Successful Response (200 OK):
// Returns a JSON array containing the list of roles based on the user's access permissions.
//
// Example Response:
//
//	[
//	    {
//	        "id": "0001",
//	        "name": "ADMIN",
//	        "description": "Full administrative access",
//	        "type": "default",
//	        "permissions": {
//	            "app_settings": ["view:own", "update:own"],
//	            "lab_books": ["view:*", "list:*", "update:own,share,review", "create:own", "delete:*"],
//	            "links": ["view:own,public", "list:own,public", "update:*", "create:*", "delete:*"],
//	            ...
//	        }
//	    },
//	    {
//	        "id": "0002",
//	        "name": "Teacher",
//	        "description": "Teacher, Professor",
//	        "type": "default",
//	        "permissions": {
//	            "app_settings": ["view:own", "update:own"],
//	            ...
//	        }
//	    },
//	    ...
//
// ❌ Error Responses:
//   - 401 Unauthorized → Missing or invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only GET is allowed).
//   - 500 Internal Server Error → Server issue or failure retrieving roles.
func HandleRoleList(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	scopes, err := ce.ScopeFetcher(pbClient, userId, casbin.PermissionConfig{
		Resources: "roles",
		Actions:   "list",
	})
	if err != nil {
		http.Error(w, "Failed to fetch user permissions", http.StatusInternalServerError)
		return
	}

	roles, err := pbClient.ListRoles(scopes, "")
	if err != nil {
		http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roles); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
