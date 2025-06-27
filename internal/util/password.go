package util

import (
	"crypto/sha1"
	"fmt"
)

const salt = "q123opdwkj9-3"

func GeneratePasswordHash(password string) string {
	hasher := sha1.New()
	hasher.Write([]byte(password))
	return fmt.Sprintf("%x", hasher.Sum([]byte(salt)))
}

func CheckPasswordHash(password string, password_hash string) bool {
	passwordHash := GeneratePasswordHash(password)
	return passwordHash == password_hash
}
