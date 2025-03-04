package tools

import (
	"mime/multipart"
	"net/http"
)

// CheckMimeType reads the first 512 bytes of a multipart.File to determine its MIME type.
//
// It returns the MIME type as a string or an error if it cannot be determined.
func CheckMimeType(file multipart.File) (string, error) {
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
