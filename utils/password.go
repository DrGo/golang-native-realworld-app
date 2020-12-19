package utils

import (
	"golang.org/x/crypto/bcrypt"
)

const numberOfTrials = 5

func HashedPassword(password string) string {
	var err error
	for i := 0; i < numberOfTrials; i++ {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err == nil {
			return string(hash)
		}
	}
	panic("HashPassword failed: " + err.Error())
}

func ValidPassword(hashedPassword string, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false
	}
	return true
}
