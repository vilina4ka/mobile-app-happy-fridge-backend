package model

import "testing"

// TestUser ...
func TestUser(t *testing.T) *User {
	return &User{
		Email:    "useremail@example.com",
		Password: "example_password",
	}
}
