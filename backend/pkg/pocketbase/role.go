package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Role struct {
	Id          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"type,omitempty"`
	Permissions interface{} `json:"permissions,omitempty"`
}

type NewRoleRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Permissions interface{} `json:"permissions"`
	Type        string      `json:"type"`
}

// ListRoles retrieves all roles from the PocketBase database.
func (pbClient *PocketBaseClient) ListRoles(fields []string, filter string) (roles []Role, err error) {
	url := fmt.Sprintf("%s/api/collections/roles/records", pbClient.BaseURL)
	// Add fields as query parameters if specified
	url += "?fields=" + strings.Join(fields, ",")

	if filter != "" {
		url += fmt.Sprintf("&filter=%s", filter)
	}

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
	var roleListResp struct {
		Page       int    `json:"page"`
		PerPage    int    `json:"perPage"`
		TotalItems int    `json:"totalItems"`
		Items      []Role `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&roleListResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return roleListResp.Items, nil
}

// CreateRole creates a new role in PocketBase.
func (pbClient *PocketBaseClient) CreateRole(role NewRoleRequest) error {
	url := fmt.Sprintf("%s/api/collections/roles/records", pbClient.BaseURL)

	data := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
		"permissions": role.Permissions,
		"type":        "custom",
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// func (pbClient *PocketBaseClient) ViewRole(roleId string) (*Role, error) {

// }

// DeleteRole deletes a role by its ID.
func (pbClient *PocketBaseClient) DeleteRole(roleId string) error {
	url := fmt.Sprintf("%s/api/collections/roles/records/%s", pbClient.BaseURL, roleId)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil

}
