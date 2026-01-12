package models

import (
	"time"
)

type User struct {
	ID           int        `json:"id"`
	UUID         string     `json:"uuid"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Password     string     `json:"-"`
	FullName     string     `json:"full_name"`
	Phone        *string    `json:"phone"`
	ProfileImage *string    `json:"profile_image"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
