package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"net/http"
)

// View User Profile
// Only users with the view:"own" permission on the "users" resource can retrieve their profile information.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `GET`
//
// ✅ Query Parameter:
//   - `id` (string, required) → The ID of the user whose profile is being requested.
//
// ✅ Successful Response (200 OK):
// Returns a JSON object containing the user's profile information.
//
// Example Response:
//
//	{
//	    "id": "user123",
//	    "name": "John Doe",
//	    "email": "johndoe@example.com",
//	    "role": "Teacher",
//	    "gender": "Male",
//	    "birthDate": "1990-05-15"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing required User ID parameter.
//   - 401 Unauthorized → Missing or invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only GET is allowed).
//   - 500 Internal Server Error → Server issue or failure retrieving user information.
func HandleUserView(w http.ResponseWriter, r *http.Request, userId string, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if userId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Extract token from header and validate it
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "view",
		Scopes:    "own",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userInfo, err := pbClient.ViewUser(userId)
	if err != nil {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userInfo)
}
