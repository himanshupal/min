package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Link struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Short     string             `json:"short"`
	Url       string             `json:"url"`
	Info      string             `json:"info"`
	CreatedBy primitive.ObjectID `json:"createdBy" bson:"createdBy"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	ExpireAt  time.Time          `json:"expireAt" bson:"expireAt"`

	CreatedAtString string `json:"-" bson:"-"`
	ExpireAtString  string `json:"-" bson:"-"`
}

type LinkError struct {
	Url      string
	Info     string
	ExpireAt string
	Password string
}

func (l Link) IsValid(savedPassword, inputPassword string) (LinkError, bool) {
	var err LinkError

	if l.Url == "" {
		err.Url = "Please provide a URL!"
	}

	if l.Info == "" {
		err.Info = "No info provided, use - to omit!"
	}

	if l.ExpireAt.Before(time.Now()) {
		err.ExpireAt = "Invalid expiration time!"
	}

	if savedPassword != inputPassword {
		err.Password = "Invalid password!"
	}

	if err.Url == "" && err.Info == "" && err.ExpireAt == "" && err.Password == "" {
		return LinkError{}, true
	}

	return err, false
}

func (l *Link) FormatTime(layout ...string) {
	// var parseLayout string = "2006-01-02 03:04:05.000 -0700 UTC"
	var parseToLayout string = "Monday, Jan 02 2006 @ 03:04 PM"

	// for _, formatString := range layout {
	// 	if formatString != "" {
	// 		parseLayout = formatString
	// 	}
	// }

	// createdAt, err := time.Parse(parseLayout, l.CreatedAt.String())
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// expireAt, err := time.Parse(parseLayout, l.ExpireAt.String())
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// l.CreatedAt = createdAt
	// l.ExpireAt = expireAt

	l.CreatedAtString = l.CreatedAt.Format(parseToLayout)
	l.ExpireAtString = l.ExpireAt.Format(parseToLayout)
}
