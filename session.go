package main

import (
	"net/http"

	uuid "github.com/gofrs/uuid"
)

func getUser(r *http.Request) user {

	c, err := r.Cookie("charlescookie")
	if err != nil {
		id, _ := uuid.NewV4()
		c = &http.Cookie{
			Name:  "charlescookie",
			Value: id.String(),
			Path:  "/",
		}

	}

	var u user
	if un, ok := dbSession[c.Value]; ok {
		u = dbUser[un]
	}
	return u
}


func alreadyloggedin(r *http.Request) bool {
	c, err := r.Cookie("charlescookie")
	if err != nil {
		return false
	}
	un := dbSession[c.Value]
	_,ok := dbUser[un]
	return ok
}
