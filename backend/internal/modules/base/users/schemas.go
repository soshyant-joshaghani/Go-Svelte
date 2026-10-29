package users

import (
	"bytes"
	"encoding/json"
)

// User is the row of table "user".
type User struct {
	ID             string
	Email          string
	IsActive       bool
	IsSuperuser    bool
	FullName       *string
	HashedPassword string
}

type UserPublic struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	IsActive    bool    `json:"is_active"`
	IsSuperuser bool    `json:"is_superuser"`
	FullName    *string `json:"full_name"`
}

func (u User) Public() UserPublic {
	return UserPublic{ID: u.ID, Email: u.Email, IsActive: u.IsActive, IsSuperuser: u.IsSuperuser, FullName: u.FullName}
}

type UsersPublic struct {
	Data  []UserPublic `json:"data"`
	Count int          `json:"count"`
}

type UserCreate struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	IsActive    *bool   `json:"is_active"`
	IsSuperuser *bool   `json:"is_superuser"`
	FullName    *string `json:"full_name"`
}

// OptString tells "absent" (leave alone) from null (clear) for full_name.
type OptString struct {
	Set   bool
	Value *string
}

func (o *OptString) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}

// UserUpdate: only the fields that were sent change.
type UserUpdate struct {
	Email       *string   `json:"email"`
	Password    *string   `json:"password"`
	FullName    OptString `json:"full_name"`
	IsActive    *bool     `json:"is_active"`
	IsSuperuser *bool     `json:"is_superuser"`
}
