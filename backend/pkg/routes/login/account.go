package login

import (
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"net/http"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login with Email & Password
// ✅ Request Body (JSON):
//
//	{
//	    "email": "user@example.com",
//	    "password": "securepassword"
//	}
//
// ✅ Successful Response (200 OK):
//
//	{
//			"status": "success",
//			"timestamp": "2025-01-30 17:23:01",
//		    "token": "your-auth-token"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Invalid JSON or missing fields
//   - 500 Internal Server Error → Server issue
func HandleAccountLogin(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient) {
	var loginData loginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := pbClient.AuthUserWithPassword(loginData.Email, loginData.Password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"token":     token,
		"timestamp": tools.Timestamp(),
	})
}
