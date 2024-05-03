package tools

import (
	"crypto/rand"
	"fmt"
)

// ChatGPT version
// generateUUID generates a random UUID (version 4) using crypto/rand
func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// Set the version (4) and variant (2 bits fixed) bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant (RFC 4122)

	// Format the UUID as a string in the standard format
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
