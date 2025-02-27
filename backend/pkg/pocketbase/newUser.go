package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type newUserResp struct {
	CollectionId   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Id             string `json:"id"`
}

type userSettingResp struct {
	CollectionId   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Id             string `json:"id"`
	Theme          string `json:"theme"`
	Language       string `json:"language"`
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
	req.Header.Set("Authorization", "Bearer "+pbClient.SuperToken)

	// Send request
	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create user: non-200 status code")
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

func (pbClient *PocketBaseClient) createDefaultSettings() (newSettingsId string, err error) {
	url := fmt.Sprintf("%s/api/collections/user_settings/records", pbClient.BaseURL)

	defaultSettings := map[string]interface{}{
		"theme":    "light",
		"language": "233",
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
	req.Header.Set("Authorization", "Bearer "+pbClient.SuperToken)

	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate user: non-200 status code")
	}

	// Parse response to extract token
	var respData userSettingResp
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respData.Id, nil
}
