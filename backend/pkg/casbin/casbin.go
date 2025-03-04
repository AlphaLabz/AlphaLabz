package casbin

import (
	"alphalabz/pkg/pocketbase"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// CasbinEnforcer struct to manage RBAC enforcement
type CasbinEnforcer struct {
	Enforcer *casbin.Enforcer
	mu       sync.Mutex // Prevents race conditions when updating policies
}

type PermissionConfig struct {
	Resources string
	Actions   string
	Scopes    string
}

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
	// Type       string                 `json:"type"`
	Permissions map[string]interface{} `json:"permissions"`
}

// InitializeCasbin initializes Casbin with provided policies (no file storage)
func InitializeCasbin(policies [][]interface{}) (*CasbinEnforcer, error) {
	// Define RBAC model
	rbacModel := `
		[request_definition]
		r = sub, obj, act, scope

		[policy_definition]
		p = sub, obj, act, scope

		[role_definition]
		g = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act && r.scope == p.scope
	`

	// Create model
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin model: %v", err)
	}

	// Use an in-memory adapter (No file storage)
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %v", err)
	}

	// Add initial policies dynamically
	for _, policy := range policies {
		_, _ = e.AddPolicy(policy...)
	}

	log.Println("Casbin Enforcer initialized successfully.")
	return &CasbinEnforcer{Enforcer: e}, nil
}

// ReloadPolicies fetches the latest policies from PocketBase and updates Casbin
func (ce *CasbinEnforcer) ReloadPolicies(pbClient *pocketbase.PocketBaseClient) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	log.Println("Reloading policies from PocketBase...")

	// Fetch latest policies from PocketBase
	policies, err := FetchPermissions(pbClient)
	if err != nil {
		return fmt.Errorf("failed to fetch policies: %v", err)
	}

	// Clear existing policies
	ce.Enforcer.ClearPolicy()

	// Add the latest policies
	for _, policy := range policies {
		_, _ = ce.Enforcer.AddPolicy(policy...)
	}

	log.Println("Casbin policies reloaded successfully.")
	return nil
}

// StartPolicyAutoReload starts a background goroutine that periodically reloads policies
func (ce *CasbinEnforcer) StartPolicyAutoReload(pbClient *pocketbase.PocketBaseClient, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := ce.ReloadPolicies(pbClient)
			if err != nil {
				fmt.Println("Error reloading policies:", err)
			}
		}
	}()
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

		for resource, actions := range role.Permissions {
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
