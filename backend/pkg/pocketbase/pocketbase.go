package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// PocketBaseClient interacts with the PocketBase HTTP API
type PocketBaseClient struct {
	BaseURL    string
	SuperToken string
	HTTPClient *http.Client
}

// NewPocketBase initializes a new PocketBase client, authenticates, and verifies the connection.
func NewPocketBase(baseURL, superuserEmail, superuserPassword string, maxRetries int, retryInterval time.Duration) (*PocketBaseClient, error) {
	client := &PocketBaseClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second}}

	// Authenticate superuser and store the token
	token, err := client.authenticateSuperuser(superuserEmail, superuserPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate superuser: %w", err)
	}
	log.Println("Successfully authenticated superuser")
	client.SuperToken = token

	// Verify PocketBase connection with retries
	for i := 0; i < maxRetries; i++ {
		err := client.CheckConnection()
		if err == nil {
			log.Println("Successfully connected to PocketBase")
			return client, nil
		}
		log.Printf("Failed to connect to PocketBase, attempt %d/%d. Retrying in %s...", i+1, maxRetries, retryInterval)
		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("failed to connect to PocketBase after %d attempts", maxRetries)
}

// authenticateSuperuser logs in the superuser and retrieves the authentication token
func (pbClient *PocketBaseClient) authenticateSuperuser(email, password string) (string, error) {
	type authWithPasswordResp struct {
		Token  string `json:"token"`
		Record struct {
			Id string `json:"id"`
		} `json:"record"`
	}

	type impersonateResp struct {
		Token string `json:"token"`
	}

	url := fmt.Sprintf("%s/api/collections/_superusers/auth-with-password", pbClient.BaseURL)

	// Data payload for authentication
	data := map[string]interface{}{
		"identity": email,
		"password": password,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate superuser: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate superuser: status %d", resp.StatusCode)
	}

	// Get auth-with-password data
	var authWPR authWithPasswordResp
	if err = json.NewDecoder(resp.Body).Decode(&authWPR); err != nil {
		return "", fmt.Errorf("failed to decode auth-with-password response: %w", err)
	}

	// Get impersonate token
	impersonateUrl := fmt.Sprintf("%s/api/collections/_superusers/impersonate/%s", pbClient.BaseURL, authWPR.Record.Id)

	type DurationPayload struct {
		Duration int `json:"duration"`
	}

	duration := DurationPayload{
		Duration: 2592000, // 30 days
	}

	impBody, err := json.Marshal(duration)
	if err != nil {
		return "", fmt.Errorf("failed to marshal duration: %w", err)
	}

	impReq, err := http.NewRequest(http.MethodPost, impersonateUrl, bytes.NewBuffer(impBody))
	if err != nil {
		return "", fmt.Errorf("failed to create impersonate request: %w", err)

	}
	impReq.Header.Add("Authorization", "Bearer "+authWPR.Token)
	impReq.Header.Set("Content-Type", "application/json")

	resp, err = pbClient.HTTPClient.Do(impReq)
	if err != nil {
		return "", fmt.Errorf("failed to impersonate superuser: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to impersonate superuser: status %d", resp.StatusCode)
	}

	var impersonateTokenResp impersonateResp
	if err = json.NewDecoder(resp.Body).Decode(&impersonateTokenResp); err != nil {
		return "", fmt.Errorf("failed to decode impersonate response: %w", err)
	}

	return impersonateTokenResp.Token, nil
}

// Check pocketbase connection is working
func (pbClient *PocketBaseClient) CheckConnection() error {
	url := fmt.Sprintf("%s/api/health", pbClient.BaseURL)

	// Create a new HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make GET request to health endpoint
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to PocketBase: %w", err)
	}
	defer resp.Body.Close()

	// Check if status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PocketBase health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (pbClient *PocketBaseClient) StartSuperTokenAutoRenew(superuserEmail, superuserPassword string) {
	interval := 24*30*time.Hour - 1*time.Hour // 30 days
	go func() {
		for {
			time.Sleep(interval)
			token, err := pbClient.authenticateSuperuser(superuserEmail, superuserPassword)
			if err != nil {
				log.Fatal("Error renewing token:", err)
				continue // Prevents goroutine from crashing
			}
			pbClient.SuperToken = token
			log.Println("SuperToken successfully renewed.")
		}
	}()
}
