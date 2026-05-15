package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	Admin    bool   `json:"admin"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Credits  int    `json:"credits"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
}

func (u *User) encryptPassword() {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	u.Password = string(hashed)
}

func (u *User) checkPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}