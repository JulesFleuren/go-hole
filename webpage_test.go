package main

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func BenchmarkBcryptHashing(b *testing.B) {

	pwd := []byte("admin")
	hash, _ := bcrypt.GenerateFromPassword(pwd, 8)
	bcrypt.CompareHashAndPassword(hash, pwd)
}
