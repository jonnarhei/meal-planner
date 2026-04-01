package models

import "golang.org/x/crypto/bcrypt"

func (u *User) SetPassword(plain string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	u.Password = hashed
	return nil
}

func (u *User) CheckPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(plain))
	return err == nil
}
