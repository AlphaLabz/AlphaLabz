package casbin

// GetRoleIDsByPermission returns a list of role IDs that have the specified permission
func (ce *CasbinEnforcer) GetRoleIDsByPermission(resource, action, scope string) ([]string, error) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	var roleIDs []string

	// Get all policies
	policies, err := ce.Enforcer.GetPolicy()
	if err != nil {
		return nil, err
	}

	// Iterate through policies and find matching ones
	for _, policy := range policies {
		if len(policy) < 4 {
			continue
		}

		roleID, obj, act, scp := policy[0], policy[1], policy[2], policy[3]

		if obj == resource && act == action && (scope == "" || scp == scope) {
			roleIDs = append(roleIDs, roleID)
		}
	}

	return roleIDs, nil
}
