package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/patrickmn/go-cache"
)

// User represents a user record from PocketBase
type User struct {
	Id        string  `json:"id,omitempty"`
	Email     string  `json:"email,omitempty"`
	Name      string  `json:"name,omitempty"`
	Avatar    string  `json:"avatar,omitempty"`
	Gender    string  `json:"gender,omitempty"`
	RoleId    string  `json:"role,omitempty"`
	SettingId string  `json:"user_settings,omitempty"`
	BirthDate string  `json:"birthdate,omitempty"`
	Expand    *Expand `json:"expand,omitempty"`
	Created   string  `json:"created,omitempty"`
	Updated   string  `json:"updated,omitempty"`
}

type Expand struct {
	Role        *Role        `json:"role,omitempty"`
	UserSetting *UserSetting `json:"userSetting,omitempty"`
}

type UserSetting struct {
	Id          string `json:"id,omitempty"`
	AppLanguage string `json:"language,omitempty"`
	Theme       string `json:"theme,omitempty"`
}

// ListUsersResponse represents the PocketBase API response for listing users
type ListUsersResponse struct {
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
	TotalItems int    `json:"totalItems"`
	Items      []User `json:"items"`
}

// ListUsers fetches the list of users
func (pbClient *PocketBaseClient) ListUsers(fields []string, expand []string, filter string) (userList []User, totalUsers int, err error) {
	url := fmt.Sprintf("%s/api/collections/users/records", pbClient.BaseURL)

	// Add fields as query parameters if specified
	url += "?fields=" + strings.Join(fields, ",")

	if len(expand) != 0 {
		url += "&expand=" + strings.Join(expand, ",")
	}

	if filter != "" {
		url += "&filter=" + filter
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Authorization header
	req.Header.Add("Authorization", "Bearer "+pbClient.SuperToken)

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("failed to fetch users: server returned status %d", resp.StatusCode)
	}

	// Parse response
	var userListResp ListUsersResponse
	err = json.NewDecoder(resp.Body).Decode(&userListResp)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return userListResp.Items, userListResp.TotalItems, nil
}

// Fetch user info from the server using userId.
// If the user info is already cached, return it immediately. Otherwise, fetch it from the server and cache it for future use.
func (pbClient *PocketBaseClient) ViewUser(userId string) (User, error) {
	// Check if user info is already cached
	userInfo, found := pbClient.UserInfoCache.Get(userId)
	if found {
		return userInfo.(User), nil
	} else {
		// Fetch user info from the server and cache it
		url := fmt.Sprintf("%s/api/collections/users/records/%s?expand=role,user_settings", pbClient.BaseURL, userId)

		// Create HTTP request and send it to the server
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return User{}, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

		// Send HTTP request and receive response from the server
		resp, err := pbClient.HTTPClient.Do(req)
		if err != nil {
			return User{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var userResponse User
			if err = json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
				return User{}, err
			}

			// Cache the fetched user info for future use
			pbClient.UserInfoCache.Set(userId, userResponse, cache.DefaultExpiration)
			return userResponse, nil
		} else {
			return User{}, fmt.Errorf("failed to fetch user info: %s", resp.Status)
		}

	}
}

// RegisterUser registers a new user in the "users" collection
func (pbClient *PocketBaseClient) NewUser(email, password, passwordConfirm, name, gender, birthDate, roleId, avatarPath string) error {
	url := fmt.Sprintf("%s/api/collections/users/records", pbClient.BaseURL)

	// Create default settings record for new user
	newSettingId, err := pbClient.createDefaultSettings()
	if err != nil {
		return fmt.Errorf("failed to create default settings: %w", err)
	}

	// Create new user record
	newUserData := map[string]interface{}{
		"email":           email,
		"password":        password,
		"passwordConfirm": passwordConfirm,
		"name":            name,
		"role":            roleId,
		"user_settings":   newSettingId,
	}

	// add content to newUserData if the content is not empty
	if gender != "" {
		newUserData["gender"] = gender
	}

	if birthDate != "" {
		newUserData["birthdate"] = birthDate
	}

	// Create request body
	body, err := json.Marshal(newUserData)
	if err != nil {
		return fmt.Errorf("failed to marshal request body, %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	// Send request
	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create user: non-200 status code")
	}

	type newUserResp struct {
		CollectionId   string `json:"collectionId"`
		CollectionName string `json:"collectionName"`
		Id             string `json:"id"`
	}

	var newUserRecord newUserResp
	if err := json.NewDecoder(resp.Body).Decode(&newUserRecord); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Uplaod user avatar
	if avatarPath != "" {
		err := pbClient.UpdateAvatar(newUserRecord.Id, avatarPath)
		if err != nil {
			return fmt.Errorf("failed to upload avatar file %w", err)
		}
	}

	return nil
}

// UpdateAvatar updates the user's profile.
func (pbClient *PocketBaseClient) UpdateProfile(userId string, newProfile User) error {
	url := fmt.Sprintf("%s/api/collections/users/records/%s", pbClient.BaseURL, userId)

	body, err := json.Marshal(newProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %w", err)
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// UpdateAvatar updates the user's avatar.
func (pbClient *PocketBaseClient) UpdateAvatar(userId, avatarPath string) error {
	url := fmt.Sprintf("%s/api/collections/users/records/%s", pbClient.BaseURL, userId)

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
	resp, err := pbClient.HTTPClient.Do(req)
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

// UpdateSettings updates the settings record for the user.
func (pbClient *PocketBaseClient) UpdateSettings(settingsId string, newSettings map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/collections/user_settings/records/%s", pbClient.BaseURL, settingsId)

	fmt.Println(url)

	body, err := json.Marshal(newSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update settings: non-200 status code")
	}
	return nil
}

// DeleteUser deletes a user by their ID.
func (pbClient *PocketBaseClient) DeleteUser(userId string) error {
	url := fmt.Sprintf("%s/api/collections/users/records/%s", pbClient.BaseURL, userId)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete user: non-204 status code")
	}

	return nil
}

// CheckUserExists checks if a user exists by their ID.
func (pbClient *PocketBaseClient) CheckUserExists(userId string) (bool, error) {
	url := fmt.Sprintf("%s/api/collections/users/records/%s", pbClient.BaseURL, userId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create check request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))
	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("failed to check user existence: non-200/404 status code")
	}
}

// ------------------------------- helper functions -------------------------------
// createDefaultSettings creates a new default settings record for the user.
func (pbClient *PocketBaseClient) createDefaultSettings() (newSettingsId string, err error) {
	url := fmt.Sprintf("%s/api/collections/user_settings/records", pbClient.BaseURL)

	defaultSettings := map[string]interface{}{
		"theme":    "light",
		"language": "en_US",
	}

	body, err := json.Marshal(defaultSettings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body, %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pbClient.SuperToken))

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate user: non-200 status code")
	}

	// Parse response to extract token
	var respData UserSetting
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respData.Id, nil
}
