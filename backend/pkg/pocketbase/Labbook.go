package pocketbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Labbook struct {
	Id            string   `json:"id"`
	Title         string   `json:"title,omitempty"`
	Description   string   `json:"description,omitempty"`
	Creator       string   `json:"creator,omitempty"`
	Reviewer      string   `json:"reviewer,omitempty"`
	ReviewStatus  string   `json:"review_status,omitempty"`
	ReviewComment string   `json:"review_comment,omitempty"`
	File          string   `json:"file,omitempty"`
	Attachments   []string `json:"attachments,omitempty"`
	ShareWith     []string `json:"share_with,omitempty"`
	CreatedAt     string   `json:"created,omitempty"`
	UpdatedAt     string   `json:"updated,omitempty"`
}

// UploadLabbook uploads a labbook and its attachments to PocketBase.
func (pbClient *PocketBaseClient) UploadLabbook(title, description, uploader, reviewer, labbookPath string, attachmentPaths []string) error {
	url := fmt.Sprintf("%s/api/collections/lab_books/records", pbClient.BaseURL)

	// Create multipart body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Open and add labbook file
	labbook, err := os.Open(labbookPath)
	if err != nil {
		return fmt.Errorf("failed to open labbook file: %w", err)
	}
	defer labbook.Close()

	labbookPart, err := writer.CreateFormFile("file", filepath.Base(labbookPath))
	if err != nil {
		return fmt.Errorf("failed to create form file for labbook: %w", err)
	}

	_, err = io.Copy(labbookPart, labbook)
	if err != nil {
		return fmt.Errorf("failed to copy labbook content: %w", err)
	}

	// Add multiple attachment files
	for _, attachmentPath := range attachmentPaths {
		attachment, err := os.Open(attachmentPath)
		if err != nil {
			return fmt.Errorf("failed to open attachment %s: %w", attachmentPath, err)
		}
		defer attachment.Close()

		// Use the same field name "attachments" for all files
		attachmentPart, err := writer.CreateFormFile("attachments", filepath.Base(attachmentPath))
		if err != nil {
			return fmt.Errorf("failed to create form file for attachment: %w", err)
		}

		_, err = io.Copy(attachmentPart, attachment)
		if err != nil {
			return fmt.Errorf("failed to copy attachment content: %w", err)
		}
	}

	// Add form fields **AFTER** adding the files
	_ = writer.WriteField("title", title)
	_ = writer.WriteField("creator", uploader)
	_ = writer.WriteField("reviewer", reviewer)
	_ = writer.WriteField("review_status", "pending")

	if description != "" {
		_ = writer.WriteField("description", description)
	}

	// Close writer to finalize form data
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Correctly set content type with the generated boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	// Send request
	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// UpdateLabbook updates a lab book record in PocketBase.
func (pbClient *PocketBaseClient) UpdateLabbook(id string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/collections/lab_books/records/%s", pbClient.BaseURL, id)

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// ListLabbooks retrieves a list of lab book records from PocketBase.
func (pbClient *PocketBaseClient) ListLabbooks(filter string, fileds []string) ([]Labbook, error) {
	url := fmt.Sprintf("%s/api/collections/lab_books/records?fields=%s&filter=(%s)", pbClient.BaseURL, strings.Join(fileds, ","), filter)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get labbooks: %w", errors.New(resp.Status))
	}

	var response struct {
		Items []Labbook `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return response.Items, nil

}

// ViewLabbook retrieves a lab book record from PocketBase.
func (pbClient *PocketBaseClient) ViewLabbook(id string, fileds []string) (Labbook, error) {
	url := fmt.Sprintf("%s/api/collections/lab_books/records/%s?fields=%s", pbClient.BaseURL, id, strings.Join(fileds, ","))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Labbook{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return Labbook{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var labbook Labbook
	if err = json.NewDecoder(resp.Body).Decode(&labbook); err != nil {
		return Labbook{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return labbook, nil
}

// ShareLabbook updates the share_with field of a lab book record in PocketBase.
func (pbClient *PocketBaseClient) ShareLabbook(id string, RecipientId string, accessList []string) error {
	url := fmt.Sprintf("%s/api/collections/lab_books/records/%s", pbClient.BaseURL, id)

	data := map[string]interface{}{
		"share_with": RecipientId,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}
