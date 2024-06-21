package relational

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap represents a map that will be stored as JSON in the database
type JSONMap map[string]interface{}

// Value converts the JSONMap to a JSON-encoded value for storing in the database
func (j JSONMap) Value() (driver.Value, error) {
	value, err := json.Marshal(j)
	return string(value), err
}

// Scan converts a JSON-encoded value from the database to a JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), j)
	case []byte:
		return json.Unmarshal(v, j)
	default:
		return errors.New("unsupported type for JSONMap")
	}
}
