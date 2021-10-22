package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Username  string             `json:"username"`
	Password  string             `json:"password"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type UserError struct {
	Username string
	Password string
}

func (u User) IsValid() (UserError, bool) {
	var err UserError

	if strings.TrimSpace(u.Username) == "" {
		err.Username = "Please choose a username!"
	} else if len(strings.TrimSpace(u.Username)) < 3 {
		err.Username = "Username must be atleast 3 characters!"
	}

	if strings.TrimSpace(u.Password) == "" {
		err.Password = "Please provide password!"
	} else if len(strings.TrimSpace(u.Password)) < 7 {
		err.Password = "Password must be atleast 7 characters!"
	}

	if err.Username == "" && err.Password == "" {
		return UserError{}, true
	}

	return err, false
}
