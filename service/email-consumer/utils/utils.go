package utils

import (
    "regexp"
    "strings"
)

// ValidEmail - checks if the given string is valid email address
func ValidEmail(email string) bool {
    if strings.TrimSpace(email) == "" {
        return false
    }

    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    matched, err := regexp.MatchString(pattern, email)
    if err != nil {
        return false
    }

    return matched
}