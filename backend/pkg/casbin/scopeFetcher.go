package casbin

import (
	"alphalabz/pkg/pocketbase"
	"fmt"
)

// ScopeFetcher fetches the scopes a user has based on their role and the permission configuration.
//
// Require Resources and Actions to be provided in PermissionConfig
func (ce *CasbinEnforcer) ScopeFetcher(pbClient *pocketbase.PocketBaseClient, userId string, permissionCfg PermissionConfig) (scopes []string, err error) {
	userRole, err := pbClient.ViewUser(userId)
	if err != nil {
		return scopes, nil
	}

	scopes, err = ce.checkPermissionScopes(userRole.RoleId, permissionCfg.Resources, permissionCfg.Actions)
	if err != nil {
		return scopes, err
	}

	return scopes, nil
}

// checkPermissionScopes retrieves all scopes a user has for a given resource and action
func (ce *CasbinEnforcer) checkPermissionScopes(roleId, resource, action string) ([]string, error) {
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
