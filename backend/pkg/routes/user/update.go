package user

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func HandlUpdateProfile(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "update",
		Scopes:    "own",
	})
	if !hasPermission || err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var updateInfo pocketbase.User
	if err := json.NewDecoder(r.Body).Decode(&updateInfo); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Validate user input
	validGender := []string{"Male", "Female", "Others", ""}
	if !tools.Contains(validGender, updateInfo.Gender) {
		http.Error(w, "Invalid gender value", http.StatusBadRequest)
		return
	}

	//	Make Sure dateOfBirth is in the yyyy-mm-dd format
	if updateInfo.BirthDate != "" {
		_, err := time.Parse("2006-01-02", updateInfo.BirthDate)
		if err != nil {
			http.Error(w, "Invalid date format, must be yyyy-mm-dd", http.StatusBadRequest)
			return
		}
	}

	// Update user account information in the database
	if err := pbClient.UpdateProfile(userId, updateInfo); err != nil {
		http.Error(w, "Failed to update user account info", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User account info updated successfully"))
}

func HandleUpdateAvatar(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Check request body content type
	if contentType := r.Header.Get("Content-Type"); contentType == "" || contentType[:19] != "multipart/form-data" {
		http.Error(w, "Invalid content type, must be multipart/form-data", http.StatusBadRequest)
		return
	}

	// Extract the token from the request header
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Verify the user's permission to update their avatar
	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "users",
		Actions:   "update",
		Scopes:    "own",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if the user has uploaded an avatar and validate it
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
		http.Error(w, "No avatar uploaded", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Failed to upload avatar", http.StatusBadRequest)
		return
	} else {
		defer file.Close()

		uploadDir := "./uploads/avatar/"
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

		mimeType, err := tools.CheckMimeType(savedFile)
		if err != nil || !allowedMimeTypes[mimeType] {
			os.Remove(filePath) // Delete the file if it's not an allowed mime type
			http.Error(w, "Invalid file format", http.StatusUnsupportedMediaType)
			return
		}
	}

	// Update the user's avatar in the database
	if err := pbClient.UpdateAvatar(userId, filePath); err != nil {
		http.Error(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Avatar updated successfully"})
	}
}
