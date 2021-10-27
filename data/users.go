package data

import (
	"fmt"

	"gorm.io/gorm"
)

var Users UserStore

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Email    string `gorm:"unique"`
	Password string
}

type UserIdentifier struct {
	ID       uint
	Username string
}

type UserPersonal struct {
	ID       uint
	Username string
	Email    string
}

type UserStore interface {
	GetUserIdentifierById(uint64) (*UserIdentifier, error)
	GetUserPersonalById(uint64) (*UserPersonal, error)
	GetUserByEmail(string) (*User, error)
	CreateUser(User) error
}

// Errors
type UserNotFoundError struct {
	ID    uint64
	Email string
}

func (e UserNotFoundError) Error() string {
	var errString string
	if e.Email != "" {
		errString = fmt.Sprintf("User with email %s was not found", e.Email)
	} else {
		errString = fmt.Sprintf("User with ID %d was not found", e.ID)
	}
	return errString
}

type UserUniquenessError struct {
	User       User
	ErrorField string
}

func (e UserUniquenessError) Error() string {
	return "User creation failed"
}

type UserCreationError struct {
	User User
	Err  error
}

func (e UserCreationError) Error() string {
	return "The user creation failed with the following error: " + e.Err.Error()
}
