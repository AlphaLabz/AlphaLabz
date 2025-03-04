package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"net/http"
)

// UserListResponse represents the response structure for the user list endpoint
type userListResponse struct {
	TotalUsers int               `json:"totalUsers"`
	Users      []pocketbase.User `json:"users"`
}

// List all users in the database
//
// ✅ Authorization:
// - Requires an `Authorization` header with a valid token.
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "totalUsers": 2,
//	    "users": [
//	        {
//	            "id": "341qctd89t52tod",
//	            "email": "test@alphalabz.net",
//	            "name": "admin",
//	            "avatar": "flask_nh12m7gyqn.jpg",
//	            "expand": {
//	                "Role": {
//	                    "id": "0001",
//	                    "name": "ADMIN"
//	                }
//	            },
//	            "created": "2025-02-06 21:18:20.471Z",
//	            "updated": "2025-02-06 21:19:08.727Z"
//	        },
//	        {
//	            "id": "1264imwwgtg65zl",
//	            "email": "test2@alphalabz.net",
//	            "name": "test_student",
//	            "expand": {
//	                "Role": {
//	                    "id": "0003",
//	                    "name": "STUDENT"
//	                }
//	            },
//	            "gender": "Others",
//	            "created": "2025-02-09 09:57:44.754Z",
//	            "updated": "2025-02-09 09:57:44.754Z"
//	        }
//	    ]
//	}
//
// ❌ Error Responses:
//   - 401 Unauthorized → Missing or Invalid Authorization token
//   - 500 Internal Server Error → Server issue
func HandleUserList(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the authorization token from the request header
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

	// Fetch user permissions based on the authorization token
	scopes, err := ce.ScopeFetcher(pbClient, userId, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "list",
	})
	if err != nil {
		http.Error(w, "Failed to fetch user permissions", http.StatusInternalServerError)
		return
	}

	userList, TotalUsers, err := pbClient.ListUsers(scopes, []string{}, "")
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	result := userListResponse{
		TotalUsers: TotalUsers,
		Users:      userList,
	}

	// Encode and send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
