package entities

import (
	"database/sql/driver"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// MultiString represents a slice of strings that can be serialized to a single string for database storage
type MultiString []string

// Scan implements the sql.Scanner interface for reading from the database
func (m *MultiString) Scan(value interface{}) error {
	if value == nil {
		*m = []string{}
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan MultiString")
	}
	*m = strings.Split(str, ",")
	return nil
}

// Value implements the driver.Valuer interface for writing to the database
func (m MultiString) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "", nil
	}
	return strings.Join(m, ","), nil
}

// GormDBDataType returns the database type for GORM
func (MultiString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite", "postgres":
		return "text"
	}
	return ""
}
