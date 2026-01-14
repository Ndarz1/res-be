package models

import "time"

type Review struct {
	ID          int       `json:"id"`
	WisataID    int       `json:"wisata_id"`
	UserID      int       `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	UserProfile *string   `json:"user_profile,omitempty"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}
