package shared

import (
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidateUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

func ValidatePassword(password string) bool {
	return len(password) >= 8
}

func SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}