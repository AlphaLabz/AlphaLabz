package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/settings"
	"alphalabz/pkg/smtp"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Invitee struct {
	Email  string `json:"email"`
	RoleId string `json:"role_id"`
}

// Invite a New User.
// Only users with the appropriate permissions can invite new users.
//
// ✅ Authorization:
// - Requires an `Authorization` header with a valid token.
// - The requesting user must have permission to create users (determined via Casbin).
//
// ✅ Request Body (JSON):
//
//	{
//	    "email": "test@example.com",
//	    "role_id": "0003" // Allowed values depend on available roles, excluding "0001"
//	}
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "link": "https://example.com/invite?token=generated-invite-token"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Invalid JSON or missing fields
//   - 401 Unauthorized → Missing or invalid Authorization token
//   - 403 Forbidden → User is not authorized to create this role
//   - 404 Not Found → Role does not exist
//   - 405 Method Not Allowed → Request method is not POST
//   - 500 Internal Server Error → Server issue or failure in generating invite link
func HandleInviteNewUser(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer, sc *smtp.SMTPClient) {
	// Constrain request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from header and validate it
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var inviteData Invitee
	if err := json.NewDecoder(r.Body).Decode(&inviteData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if inviteData.Email == "" || inviteData.RoleId == "" {
		http.Error(w, "Email and name are required", http.StatusBadRequest)
		return
	}

	// Get available roles
	roles, err := pbClient.ListRoles([]string{"id", "name", "type"}, "")
	if err != nil {
		http.Error(w, "Failed to get available roles", http.StatusInternalServerError)
		return
	}

	// Check if role exists
	roleExists := false
	for _, role := range roles {
		if role.Id == inviteData.RoleId {
			roleExists = true
			break
		}
	}
	if !roleExists {
		http.Error(w, "Role not found", http.StatusBadRequest)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Grant user scopes
	scopes, err := ce.ScopeFetcher(pbClient, userId, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "create",
	})
	if err != nil {
		http.Error(w, "Failed to fetch user scopes", http.StatusInternalServerError)
		return
	}

	// Check if user is authorized to create this role
	if inviteData.RoleId == "0001" {
		http.Error(w, "Unauthorized to create this role", http.StatusForbidden)
		return
	} else {
		var inviteeData = Invitee{
			RoleId: inviteData.RoleId,
			Email:  inviteData.Email,
		}
		if tools.Contains(scopes, "*") {
			// Allow user to create all other roles
			sendInviteResponse(w, inviteeData)
			return
		}

		if tools.Contains(scopes, inviteData.RoleId) {
			// Allow user to create this role
			sendInviteResponse(w, inviteeData)
			return
		}

		if !tools.Contains(scopes, inviteData.RoleId) && !tools.Contains(scopes, "*") {
			http.Error(w, "Unauthorized to create this role", http.StatusForbidden)
			return
		}

	}

}

func sendInviteResponse(w http.ResponseWriter, invitee Invitee) {
	inviteLink, err := generateInvitation(invitee)
	if err != nil {
		http.Error(w, "Failed to generate invitation link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"invite_link": inviteLink})
}

func generateInvitation(invitee Invitee) (inviteLink string, err error) {
	// Get JWT secret from settings
	settings, err := settings.LoadSettings("settings.yml")
	if err != nil {
		return "", fmt.Errorf("failed to load settings")
	} else {
		// Get invite JWT secret from settings
		inviteSecret := settings.JWTSecret
		if inviteSecret == "" {
			return "", fmt.Errorf("invite secret not set")
		}

		// Generate JWT token
		secretKey := []byte(inviteSecret)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256,
			jwt.MapClaims{
				"email":   invitee.Email,
				"role_id": invitee.RoleId,
				"exp":     time.Now().Add(time.Hour * 24).Unix(),
			})

		tokenString, err := token.SignedString(secretKey)
		if err != nil {
			return "", err
		}

		// Format the invite link
		inviteLink = fmt.Sprintf("%s/invite?token=%s", settings.AppUrl, tokenString)
		return inviteLink, nil
	}
}
