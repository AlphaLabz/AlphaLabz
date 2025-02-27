package pocketbase

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func (pbClient *PocketBaseClient) UpdateAvatar(newUserRecordId, avatarPath string) error {
	url := fmt.Sprintf("%s/api/collections/users/records/%s", pbClient.BaseURL, newUserRecordId)

	// Open avatar file
	file, err := os.Open(avatarPath)
	if err != nil {
		return fmt.Errorf("failed to open avatar file: %w", err)
	}
	defer file.Close()

	// Create multipart body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file for avatar data
	part, err := writer.CreateFormFile("avatar", filepath.Base(avatarPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close writer after we are done with it
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create HTTP request with multipart body and headers set up for PATCH method
	req, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+pbClient.SuperToken)

	// Do the request and get response back from server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check if request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to upload avatar, status code: %d", resp.StatusCode)
	}

	return nil
}
