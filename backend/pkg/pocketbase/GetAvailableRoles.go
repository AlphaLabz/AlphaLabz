package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Role struct {
	Id          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Permissions interface{} `json:"permission"`
}

type ListRolesResponse struct {
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
	TotalItems int    `json:"totalItems"`
	Items      []Role `json:"items"`
}

// Get all available roles in the database
func (pbClient *PocketBaseClient) GetAvailableRoles(fields []string) (roles []Role, err error) {
	url := fmt.Sprintf("%s/api/collections/roles/records", pbClient.BaseURL)
	// Add fields as query parameters if specified
	url += "?fields=" + strings.Join(fields, ",")

	fmt.Println(url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+pbClient.SuperToken)

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close() // Ensure body is closed after reading

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch roles: received status code %d", resp.StatusCode)
	}

	// Parse JSON response
	var roleListResp ListRolesResponse
	err = json.NewDecoder(resp.Body).Decode(&roleListResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return roleListResp.Items, nil
}
