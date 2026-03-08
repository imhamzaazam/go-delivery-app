package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/horiondreher/go-web-api-boilerplate/internal/domain/domainerr"
	"golang.org/x/crypto/bcrypt"
)

const NoPasswordPrefix = "!!NOLOGIN!!"

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

// HashPasswordOrNoop hashes the password if non-empty. If empty, returns a
// random sentinel value that can never match a bcrypt comparison, effectively
// preventing login for the actor.
func HashPasswordOrNoop(password string) (string, *domainerr.DomainError) {
	if strings.TrimSpace(password) == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", domainerr.NewDomainError(http.StatusInternalServerError, domainerr.InternalError, "failed to generate noop hash", err)
		}
		return NoPasswordPrefix + hex.EncodeToString(b), nil
	}
	return HashPassword(password)
}

func CheckPassword(password string, hashedPassword string) *domainerr.DomainError {
	if strings.HasPrefix(hashedPassword, NoPasswordPrefix) {
		return domainerr.NewDomainError(http.StatusUnauthorized, domainerr.UnauthorizedError, "login not allowed for this account", fmt.Errorf("login not allowed for passwordless account"))
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return domainerr.MatchHashError(err)
	}

	return nil
}
