package service

import (
	"errors"
	"unicode"
	"unicode/utf8"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
)

const PasswordPolicyMessage = "Password must be 8-128 characters and include uppercase, lowercase, numeric, and special characters"

var ErrInvalidPassword = errors.New(PasswordPolicyMessage)

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength || length > maxPasswordLength {
		return ErrInvalidPassword
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrInvalidPassword
	}
	return nil
}
