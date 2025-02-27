package casbin

import (
	"alphalabz/pkg/pocketbase"
	"fmt"
	"log"
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
