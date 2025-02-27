package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserRole struct {
	UserId string `json:"id"`
	Name   string `json:"name"`
	RoleId string `json:"role"`
}

func (pbClient *PocketBaseClient) FetchUserRole(userJwt string) (UserRole, error) {
	url := fmt.Sprintf("%s/api/role/%s", pbClient.BaseURL, userJwt)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return UserRole{}, err
	}

	// Add autorization header
	req.Header.Set("Authorization", "Bearer "+pbClient.SuperToken)

	// Send request
	resp, err := pbClient.HTTPClient.Do(req)
	if err != nil {
		return UserRole{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserRole{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var userRoleResponse UserRole
	if err := json.NewDecoder(resp.Body).Decode(&userRoleResponse); err != nil {
		return UserRole{}, err
	}

	return userRoleResponse, nil
}
