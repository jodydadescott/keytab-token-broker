package policy

// Decision ...
type Decision struct {
	Auth       bool
	Principals []string
}

// HasPrincipal Returns true if principal is present in entity
func (t *Decision) HasPrincipal(principal string) bool {
	if principal == "" {
		return false
	}
	for _, s := range t.Principals {
		if s == principal {
			return true
		}
	}
	return false
}
