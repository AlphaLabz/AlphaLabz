package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/settings"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Sign Up a New User
//
// ✅ Description:
// - Allows a new user to register in the system using a form submission.
//
// ✅ Authorization:
// - Requires an `Authorization` header with a valid token.
//
// ✅ Request Body:
// - `Content-Type: multipart/form-data`
// - Fields:
//   - `token` (string, required) → JWT token for authentication.
//   - `username` (string, required) → The desired username.
//   - `password` (string, required) → The password for the account.
//   - `passwordConfirm` (string, required) → Must match `password`.
//   - `dateOfBirth` (string, optional) → Format: yyyy-mm-dd.
//   - `gender` (string, optional) → Must be one of: "Male", "Female", "Others".
//   - `avatar` (file, optional) → Allowed formats: JPEG, JPG, PNG, GIF, HEIC, HEIF, WEBP, SVG.
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "message": "User created successfully"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing required fields, invalid password confirmation, or incorrect format.
//   - 401 Unauthorized → Missing or Invalid Authorization token.
//   - 415 Unsupported Media Type → Avatar file format is not allowed.
//   - 500 Internal Server Error → Server issue or file saving error.
func HandleSignUp(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check request body content type
	if contentType := r.Header.Get("Content-Type"); contentType == "" || contentType[:19] != "multipart/form-data" {
		http.Error(w, "Invalid content type, must be multipart/form-data", http.StatusBadRequest)
		return
	}

	// Constrain request size
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// form data
	token := r.FormValue(("token"))
	username := r.FormValue("username")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("passwordConfirm")
	dateOfBirth := r.FormValue("dateOfBirth") // yyyy-mm-dd
	gender := r.FormValue("gender")           // Male, Female, Others

	// Ensure all necessary fields are not empty
	if token == "" || username == "" || password == "" || passwordConfirm == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Parse JWT token
	roleId, _, email, err := parseJWT(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Make sure the password and passwordConfirm match
	if password != passwordConfirm {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	//	Make Sure dateOfBirth is in the yyyy-mm-dd format
	if dateOfBirth != "" {
		_, err := time.Parse("2006-01-02", dateOfBirth)
		if err != nil {
			http.Error(w, "Invalid date format, must be yyyy-mm-dd", http.StatusBadRequest)
			return
		}
	}

	//	Make Sure Gender is Male, Female or Others
	if gender != "Male" && gender != "Female" && gender != "Others" && gender != "" {
		http.Error(w, "Invalid gender", http.StatusBadRequest)
		return
	}

	// Obtain the img that upload from the user
	var allowedMimeTypes = map[string]bool{
		"image/jpeg":    true,
		"image/jpg":     true,
		"image/png":     true,
		"image/gif":     true,
		"image/heic":    true,
		"image/heif":    true,
		"image/webp":    true,
		"image/svg+xml": true,
	}

	// Check if the user has uploaded an avatar and validate it
	var filePath string
	file, handler, err := r.FormFile("avatar")
	if err == http.ErrMissingFile {
		filePath = ""
	} else if err != nil {
		http.Error(w, "Failed to upload avatar", http.StatusBadRequest)
		return
	} else {
		defer file.Close()

		uploadDir := "./uploads/"
		os.MkdirAll(uploadDir, 0755)
		filePath = uploadDir + fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)

		// Save the file to disk
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// Check the mime type of the uploaded file
		savedFile, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Failed to open saved avatar", http.StatusInternalServerError)
			return
		}
		defer savedFile.Close()

		mimeType, err := checkMimeType(savedFile)
		if err != nil || !allowedMimeTypes[mimeType] {
			os.Remove(filePath) // Delete the file if it's not an allowed mime type
			http.Error(w, "Invalid file format", http.StatusUnsupportedMediaType)
			return
		}
	}

	// Regist new user
	if err = pbClient.NewUser(email, password, passwordConfirm, username, gender, dateOfBirth, roleId, filePath); err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Response
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func parseJWT(tokenString string) (roleId, roleName, email string, err error) {
	claims := jwt.MapClaims{}

	tokenParsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the token signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Load secret key from settings
		settingsData, err := settings.LoadSettings("settings.yml")
		if err != nil {
			return nil, fmt.Errorf("failed to load settings")
		}

		// Return secret key for verification
		return []byte(settingsData.JWTSecret), nil
	})

	if err != nil || !tokenParsed.Valid {
		return "", "", "", fmt.Errorf("invalid token")
	}

	// Extract claims with type assertion
	var ok bool
	if email, ok = claims["email"].(string); !ok {
		return "", "", "", fmt.Errorf("invalid token: missing email claim")
	}
	if roleId, ok = claims["role_id"].(string); !ok {
		return "", "", "", fmt.Errorf("invalid token: missing role_id claim")
	}
	if roleName, ok = claims["role_name"].(string); !ok {
		return "", "", "", fmt.Errorf("invalid token: missing role_name claim")
	}

	return roleId, roleName, email, nil
}

func checkMimeType(file multipart.File) (string, error) {
	// Read 512 bit of the file
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}
	// reset
	file.Seek(0, 0)
	mimeType := http.DetectContentType(buf)
	return mimeType, nil
}
