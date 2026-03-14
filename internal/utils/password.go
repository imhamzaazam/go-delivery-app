package utils

import (
	"crypto/rand"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/horiondreher/go-web-api-boilerplate/internal/core/domainerr"
)

type HashError struct {
	msg string
}

func (e *HashError) Error() string {
	return e.msg
}

func HashPassword(password string) (string, *domainerr.DomainError) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", domainerr.NewDomainError(http.StatusInternalServerError, domainerr.InternalError, err.Error(), err)
	}

	return string(hashedPassword), nil
}

// HashPasswordOrNoop hashes the password if non-empty. If empty, it hashes a
// random value so passwordless accounts still cannot log in with any password.
func HashPasswordOrNoop(password string) (string, *domainerr.DomainError) {
	if strings.TrimSpace(password) == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", domainerr.NewDomainError(http.StatusInternalServerError, domainerr.InternalError, "failed to generate noop hash", err)
		}
		hashedPassword, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
		if err != nil {
			return "", domainerr.NewDomainError(http.StatusInternalServerError, domainerr.InternalError, err.Error(), err)
		}
		return string(hashedPassword), nil
	}
	return HashPassword(password)
}

func CheckPassword(password string, hashedPassword string) *domainerr.DomainError {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return domainerr.MatchHashError(err)
	}

	return nil
}
