package utils

import (
	"fmt"
	"pipec-backend/models"
	"strings"
)

type PipeAddress = models.PipeAddress

// ParseEmail parses an email address into username and domain
func ParseEmail(email string) (*PipeAddress, error) {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	username := strings.TrimSpace(parts[0])
	domain := strings.TrimSpace(parts[1])

	if username == "" || domain == "" {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	return &PipeAddress{
		Username: username,
		Domain:   domain,
		Raw:      email,
	}, nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	_, err := ParseEmail(email)
	return err
}

func GroupEmailsByDomain(recipients []PipeAddress) map[string][]PipeAddress {
	grouped := make(map[string][]PipeAddress)
	for _, r := range recipients {
		grouped[r.Domain] = append(grouped[r.Domain], r)
	}
	return grouped
}

func BatchParseEmails(emails []string) ([]PipeAddress, error) {
	var addresses []PipeAddress
	for _, email := range emails {
		addr, err := ParseEmail(email)
		if err != nil {
			return nil, fmt.Errorf("error parsing email '%s': %w", email, err)
		}
		addresses = append(addresses, *addr)
	}
	return addresses, nil
}
