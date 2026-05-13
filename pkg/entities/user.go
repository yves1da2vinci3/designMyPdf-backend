package entities

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserName    string  `json:"user_name"`
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	FirebaseUID *string `json:"firebase_uid,omitempty" gorm:"uniqueIndex"`
	Session     Session `json:"session"`
	Namespace   []Namespace `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Keys        []Key       `json:"keys" gorm:"foreignKey:UserID"`
}
