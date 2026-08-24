package database

import (
	"fmt"
	"regexp"
)

// validSchemaName is a compiled regex that enforces safe PostgreSQL schema names.
// Only lowercase letters, digits, and underscores are allowed, 3–63 characters long.
var validSchemaName = regexp.MustCompile(`^[a-z0-9_]{3,63}$`)

// ValidateSchemaName returns an error if the given schema name is not safe to use
// in a PostgreSQL SET search_path statement. This prevents SQL injection through
// user-supplied subdomain values.
func ValidateSchemaName(name string) error {
	if name == "" {
		return fmt.Errorf("schema name must not be empty")
	}
	if !validSchemaName.MatchString(name) {
		return fmt.Errorf("invalid schema name %q: must be 3–63 chars, lowercase letters, digits, and underscores only", name)
	}
	return nil
}
