package database

import (
	"fmt"
	"regexp"
)

// validSchemaName is a compiled regex that enforces safe PostgreSQL schema names.
// Only lowercase letters (a–z) are allowed, 3–63 characters long.
// Numbers, underscores, hyphens, and special characters are NOT permitted.
var validSchemaName = regexp.MustCompile(`^[a-z]{3,63}$`)

// ValidateSchemaName returns an error if the given schema name is not safe to use
// in a PostgreSQL SET search_path statement. This prevents SQL injection through
// user-supplied subdomain values.
func ValidateSchemaName(name string) error {
	if name == "" {
		return fmt.Errorf("schema name must not be empty")
	}
	if !validSchemaName.MatchString(name) {
		return fmt.Errorf("invalid subdomain %q: must be 3–63 lowercase letters only (a–z), no numbers or special characters", name)
	}
	return nil
}
