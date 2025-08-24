package models

import "time"

//for USER
type User struct {
	ID             string    `json:"id"`              // Unique identifier for the user
	Name           string    `json:"name"`            // Name of user
	Email          string    `json:"email"`           //Email address of user
	Phone          string    `json:"phone"`           //Phone number of user
	Password       string    `json:"password"`        // Password of user
	PasswordHash   string    `json:"-"`               //Hashed password of user
	MembershipTier string    `json:"membership_tier"` //membership tier status of user
	CreatedAt      time.Time `json:"created_at"`      //time of account creation
}
