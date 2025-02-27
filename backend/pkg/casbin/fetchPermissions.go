package casbin

import (
	"alphalabz/pkg/pocketbase"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type PermissionRespond struct {
	Items      []RolePermission `json:"items"`
	Page       int              `json:"page"`
	PerPage    int              `json:"perPage"`
	TotalItems int              `json:"totalItems"`
	TotalPages int              `json:"totalPages"`
}

type RolePermission struct {
	// CollectionId   string                 `json:"collectionId"`
	// CollectionName string                 `json:"collectionName"`
	// Description string                 `json:"description"`
	Id string `json:"id"`
	// Name        string                 `json:"name"`
	Permission map[string]interface{} `json:"permission"`
}

// FetchPermissions fetch the latest permissons settings from the database and save as a local csv
func FetchPermissions(pbClient *pocketbase.PocketBaseClient) ([][]interface{}, error) {
	url := fmt.Sprintf("%s/api/collections/roles/records?perPage=99", pbClient.BaseURL)

	// Construct request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+pbClient.SuperToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Parse response
	var respData PermissionRespond
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return nil, err
	}

	// Convert to Casbin policies ([][]interface{})
	return convertCasbinFormat(respData.Items), nil
}

// ConvertCasbinFormat converts RolePermissions into Casbin policy rules
func convertCasbinFormat(permissions []RolePermission) [][]interface{} {
	var casbinPolicies [][]interface{}

	for _, role := range permissions {
		roleID := role.Id // Role ID as Casbin "sub"

		for resource, actions := range role.Permission {
			actionList, ok := actions.([]interface{})
			if !ok {
				continue
			}

			for _, action := range actionList {
				actionStr, ok := action.(string)
				if !ok {
					continue
				}

				actionParts := strings.Split(actionStr, ":") // Split action & scope
				actionType := actionParts[0]                 // "view", "update", etc.
				scope := "all"                               // Default scope

				if len(actionParts) > 1 {
					scopes := strings.Split(actionParts[1], ",") // Handle multiple scopes

					for _, s := range scopes {
						casbinPolicies = append(casbinPolicies, []interface{}{
							roleID, resource, actionType, strings.TrimSpace(s),
						})
					}
					continue
				}

				// If there's only one scope, add normally
				casbinPolicies = append(casbinPolicies, []interface{}{
					roleID, resource, actionType, scope,
				})
			}
		}
	}

	return casbinPolicies
}
