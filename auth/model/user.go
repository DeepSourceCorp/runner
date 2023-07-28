package model

import "encoding/json"

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Login    string `json:"login"`
	Provider string `json:"provider"`
}

func (u *User) Claims() map[string]interface{} {
	return map[string]interface{}{
		"id":       u.ID,
		"name":     u.Name,
		"email":    u.Email,
		"login":    u.Login,
		"provider": u.Provider,
	}
}

func (u *User) String() string {
	s, _ := json.Marshal(u)
	return string(s)
}
