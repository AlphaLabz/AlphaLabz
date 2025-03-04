package labbook

import (
	"alphalabz/pkg/casbin"
	"alphalabz/pkg/pocketbase"
	"alphalabz/pkg/settings"
	"alphalabz/pkg/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Upload a Lab Book
// Only users with the create:"own" permission on the "labbook" resource can upload a lab book.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ Request Body: `Content-Type: multipart/form-data`
//
// - Fields:
//   - `title ` (string, required) → The title of the lab book.
//   - `description` (string, optional) → A brief description of the lab book.
//   - `reviwerId`(string, required) → The reviewer of the lab book.
//   - `file` (file, required) → Allowed formats: PDF, DOCX, PPTX, XLSX, DOC, XLS, PPT, MP4, PNG, MKV, JPG, MP3.
//
// ✅ Successful Response (200 OK):
//
//	{
//	    "message": "Lab book uploaded successfully"
//	}
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing required fields or invalid reqest body format.
//   - 401 Unauthorized → Missing or Invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method.
//   - 415 Unsupported Media Type →  Invalid file format.
//   - 500 Internal Server Error → Server issue or file saving error.
func HandleLabBookUpload(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	// Check if the request method is POST.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := settings.LoadSettings("settings.yml")
	if err != nil {
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	// Constrain form size
	r.ParseMultipartForm(settings.MaxLabbookSize << 20)

	// Get request authorization header
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Verify user permission using Casbin enforcer
	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "create",
		Scopes:    "own",
	})
	if err != nil || !hasPermission {
		http.Error(w, "Failed to verify permission", http.StatusInternalServerError)
		return
	}

	// Check if the request content type is multipart/form-data.
	if contentType := r.Header.Get("Content-Type"); contentType == "" || contentType[:19] != "multipart/form-data" {
		http.Error(w, "Invalid content type, must be multipart/form-data", http.StatusBadRequest)
		return
	}

	// Extract form text values.
	title := r.FormValue("title")
	description := r.FormValue("description")
	reviewerId := r.FormValue("reviewerId")

	if title == "" || reviewerId == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Handle lab book upload.
	var filePath string
	file, handler, err := r.FormFile("file")
	if err == http.ErrMissingFile {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	} else {
		defer file.Close()

		var allowedMimeTypes = map[string]bool{
			"application/pdf": true, // .pdf
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true, // .docx
			"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx
			"application/msword":            true, // .doc
			"application/vnd.ms-powerpoint": true, // .ppt
		}

		uploadDir := fmt.Sprintf("./uploads/labbook/%d/", time.Now().Unix())
		os.MkdirAll(uploadDir, 0755)
		filePath = uploadDir + handler.Filename

		// Save the file to disk
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		// Open the saved file to check its mime type
		savedFile, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Failed to open saved file", http.StatusInternalServerError)
			return
		}
		defer savedFile.Close()

		// Check if the uploaded file is one of the allowed mime types
		mimeType, err := tools.CheckMimeType(savedFile)
		if err != nil || !allowedMimeTypes[mimeType] {
			os.Remove(filePath) // Delete the file if it's not an allowed mime type
			http.Error(w, "Invalid file format", http.StatusUnsupportedMediaType)
			return
		}
	}

	// Handle multiple attachments
	var attachmentPaths []string
	if r.MultipartForm != nil {
		attachments := r.MultipartForm.File["attachments"] // Retrieve multiple files

		for _, fileHeader := range attachments {
			attachment, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Failed to retrieve attachment", http.StatusInternalServerError)
				return
			}
			defer attachment.Close()

			// Allowed MIME types
			allowedMimeTypes := map[string]bool{
				"application/pdf": true,
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true, // .docx
				"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true, // .xlsx
				"application/msword":            true, // .doc
				"application/vnd.ms-excel":      true, // .xls
				"application/vnd.ms-powerpoint": true, // .ppt
				"image/jpeg":                    true, // .jpg
				"image/jpg":                     true,
				"image/png":                     true, // .png
				"image/gif":                     true, // .gif
				"video/mp4":                     true, // .mp4
				"video/x-matroska":              true, // .mkv
				"audio/mpeg":                    true, // .mp3
			}

			// Detect the MIME type of the uploaded file
			mimeType, err := tools.CheckMimeType(attachment)
			if err != nil || !allowedMimeTypes[mimeType] {
				http.Error(w, fmt.Sprintf("Invalid file format for %s", fileHeader.Filename), http.StatusUnsupportedMediaType)
				return
			}

			// Save the attachment
			attachmentDir := fmt.Sprintf("./uploads/labbook/%d/attachments/", time.Now().Unix())
			os.MkdirAll(attachmentDir, 0755)
			attachmentPath := attachmentDir + fileHeader.Filename

			dst, err := os.Create(attachmentPath)
			if err != nil {
				http.Error(w, "Failed to save attachment", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			_, err = io.Copy(dst, attachment)
			if err != nil {
				http.Error(w, "Error saving attachment", http.StatusInternalServerError)
				return
			}

			// Append the saved attachment path
			attachmentPaths = append(attachmentPaths, attachmentPath)
		}
	}

	// Get user details from request body
	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Pass attachments to UploadLabbook
	err = pbClient.UploadLabbook(title, description, userId, reviewerId, filePath, attachmentPaths)
	if err != nil {
		http.Error(w, "Failed to upload labbook", http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Labbook uploaded successfully",
		"content": map[string]string{
			"title":       title,
			"description": description,
			"userId":      userId,
			"reviewerId":  reviewerId,
		},
	})
}

// View Lab Book Upload History
// Only users with the view:"own" permission on the "lab_books" resource can retrieve their lab book upload history.
//
// ✅ Authorization:
// Requires an `Authorization` header with a valid token.
//
// ✅ HTTP Method: `GET`
//
// ✅ Successful Response (200 OK):
// Returns the list of lab book upload history for the authenticated user in JSON format.
//
// ❌ Error Responses:
//   - 400 Bad Request → Missing or invalid request parameters.
//   - 401 Unauthorized → Missing or invalid Authorization token.
//   - 403 Forbidden → User does not have the required permissions.
//   - 405 Method Not Allowed → Invalid HTTP method (only GET is allowed).
//   - 500 Internal Server Error → Server issue or database retrieval error.
func HandleLabbookUploadHistory(w http.ResponseWriter, r *http.Request, pbClient *pocketbase.PocketBaseClient, ce *casbin.CasbinEnforcer) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rawToken, err := tools.TokenExtractor(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if the user has permission to review lab books
	hasPermission, _, err := ce.VerifyJWTPermission(pbClient, rawToken, casbin.PermissionConfig{
		Resources: "lab_books",
		Actions:   "view",
		Scopes:    "own"})
	if err != nil || !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userId, err := tools.GetUserIdFromJWT(rawToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	uploadHistory, err := pbClient.ListLabbooks(fmt.Sprintf("creator='%s'", userId), []string{"*"})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get lab book upload history", http.StatusInternalServerError)
		return
	}

	// Return the lab book upload history as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploadHistory)
}
