package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string // Может быть пустой
	Role         RoleType
}

type RoleType string

const (
	RoleEmployee  RoleType = "employee"
	RoleModerator RoleType = "moderator"
)

func (r RoleType) Valid() bool {
	switch r {
	case RoleEmployee, RoleModerator:
		return true
	}

	return false
}

func (r RoleType) String() string {
	return string(r)
}

type Token string // JWT token
