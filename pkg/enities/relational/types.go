package relational

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MultiString []string

func (MultiString) GormDBDataType(db *gorm.DB, field *schema.Field) string {

	// returns different database type based on driver name
	switch db.Dialector.Name() {
	case "mysql", "sqlite", "postgresql":
		return "text"
	}
	return ""
}
