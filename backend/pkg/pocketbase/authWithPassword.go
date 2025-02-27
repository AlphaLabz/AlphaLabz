package pocketbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// AuthenticateUser authenticates a user and returns their token
func (pbClient *PocketBaseClient) AuthUserWithPassword(email, password string) (string, error) {
	url := fmt.Sprintf("%s/api/collections/users/auth-with-password", pbClient.BaseURL)

	// Data payload for authentication
	data := map[string]interface{}{
		"identity": email,
		"password": password,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to authenticate user: non-200 status code")
	}

	// Parse response to extract token
	var respData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respData.Token, nil
}
