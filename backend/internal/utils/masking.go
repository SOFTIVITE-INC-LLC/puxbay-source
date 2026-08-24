package utils

import (
	"encoding/json"

	"github.com/softivite/puxbay/internal/middleware"
)

// MaskCollection takes any data (struct or slice of structs), marshals it to JSON,
// unmarshals it into map[string]interface{}, applies FieldRBAC, and returns it.
func MaskCollection(data interface{}, role string, permissions []string) interface{} {
	b, err := json.Marshal(data)
	if err != nil {
		return data // fallback if marshal fails
	}

	// Try unmarshaling as a slice of maps
	var sliceOfMaps []map[string]interface{}
	if err := json.Unmarshal(b, &sliceOfMaps); err == nil {
		for i := range sliceOfMaps {
			sliceOfMaps[i] = middleware.FieldRBAC(role, permissions, sliceOfMaps[i])
		}
		return sliceOfMaps
	}

	// Try unmarshaling as a single map
	var singleMap map[string]interface{}
	if err := json.Unmarshal(b, &singleMap); err == nil {
		return middleware.FieldRBAC(role, permissions, singleMap)
	}

	return data // fallback
}
