package domain

import (
	"crypto/rand"
	"crypto/sha512"
)

const saltSize = 32

func GetSalt() ([]byte, error) {
	passwordSalt := make([]byte, saltSize)
	_, err := rand.Read(passwordSalt)
	return passwordSalt, err
}

func GetHash512(password string, salt []byte) ([]byte, error) {
	var passwordHash []byte
	sha512Hasher := sha512.New()
	if _, err := sha512Hasher.Write(append([]byte(password), salt...)); err != nil {
		return passwordHash, err
	}
	passwordHash = sha512Hasher.Sum(nil)
	return passwordHash, nil
}
