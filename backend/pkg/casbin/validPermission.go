package casbin

import (
	"fmt"
)

// CheckPermission validates user actions using Casbin
func (ce *CasbinEnforcer) VerifyPermission(roleId, resource, action, scope string) (bool, error) {
	if ce.Enforcer == nil {
		return false, fmt.Errorf("casbin Enforcer is not initialized")
	}

	ok, err := ce.Enforcer.Enforce(roleId, resource, action, scope)
	if err != nil {
		fmt.Println("Error enforcing policy:", err)
		return false, err
	}
	return ok, nil
}

// CheckPermissionScopes retrieves all scopes a user has for a given resource and action
func (ce *CasbinEnforcer) CheckPermissionScopes(roleId, resource, action string) ([]string, error) {
	if ce.Enforcer == nil {
		return nil, fmt.Errorf("casbin Enforcer is not initialized")
	}

	// Retrieve all policies in Casbin
	allPolicies, err := ce.Enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve policies: %v", err)
	}

	// Store valid scopes for this user/resource/action
	var scopes []string

	// Iterate through all Casbin policies
	for _, policy := range allPolicies {
		// Policy format: [roleId, resource, action, scope]
		if len(policy) == 4 && policy[0] == roleId && policy[1] == resource && policy[2] == action {
			scopes = append(scopes, policy[3]) // Collect the allowed scopes
		}
	}

	// If no scopes were found, return an error
	if len(scopes) == 0 {
		return nil, fmt.Errorf("no scopes found for user role: %s, resource: %s, action: %s", roleId, resource, action)
	}

	return scopes, nil
}
