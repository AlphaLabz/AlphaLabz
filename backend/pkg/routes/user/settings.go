package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Update User Settings
// Only users with the update:"own" permission on the "users" resource can update their personal settings.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `PATCH`
//
// ✅ Request Body: `Content-Type: application/json`
// - Fields:
//   - `AppLanguage` (string, required) → The preferred application language (must be a valid language code).
//   - `Theme` (string, required) → The UI theme preference; allowed values: `"dark"`, `"light"`.
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "message": "Settings updated successfully"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing required fields, invalid theme value, or unsupported language code.
//   - 401 Unauthorized → Missing or Invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only PATCH is allowed).
//   - 500 Internal Server Error → Server issue, failure retrieving user data, or updating settings.
func HandlUpdateSettings(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Constrain request method
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from header and validate it
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hasPermmission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "update",
		Scopes:    "own",
	})
	if err != nil || !hasPermmission {
		fmt.Println(hasPermmission)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var updateSettings pocketbase.UserSetting
	if err := json.NewDecoder(r.Body).Decode(&updateSettings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// If AppLanguage or Theme is empty, return an error response
	if updateSettings.AppLanguage == "" || updateSettings.Theme == "" {
		http.Error(w, "Invalid request body fields", http.StatusBadRequest)
		return
	}

	// Check if the theme is valid
	if updateSettings.Theme != "dark" && updateSettings.Theme != "light" {
		http.Error(w, "Invalid theme", http.StatusBadRequest)
		return
	}

	// Get available languages from the csv file
	languages, err := readLanguagesFromCSV("./appLanguages.csv")
	if err != nil {
		http.Error(w, "Error reading languages from CSV", http.StatusInternalServerError)
		return
	}

	// Check if the language is valid
	if _, ok := languages[updateSettings.AppLanguage]; !ok {
		http.Error(w, "Invalid language code", http.StatusBadRequest)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
		return
	}

	// Get user's settings Id
	userInfo, err := pbClient.ViewUser(userId)
	if err != nil {
		http.Error(w, "Error retrieving user info", http.StatusInternalServerError)
		return
	}

	// Update user settings in the database
	if err = pbClient.UpdateSettings(userInfo.SettingId, map[string]interface{}{
		"language": updateSettings.AppLanguage,
		"theme":    updateSettings.Theme}); err != nil {
		http.Error(w, "Error updating user settings", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Settings updated successfully",
	})

}

// ReadLanguagesFromCSV reads language settings from a CSV file and returns a map
func readLanguagesFromCSV(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	languages := make(map[string]string)
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(record) < 3 {
			continue // Skip invalid rows
		}
		languages[record[1]] = record[2] // Map code to display name
	}

	return languages, nil
}
