package main

import (
	"fmt"
	"strings"
)

const (
	minDataLen = 1000

	mandatoryTag = "html>"
	errReply     = "err"
)

var (
	ErrDataLack  = fmt.Errorf("data len must be at least %d characters", minDataLen)
	ErrNoHeadTag = fmt.Errorf("data must contain tag: %s", mandatoryTag)
)

// ValidateReply performs some checks on data to make sure it isn't just some garbage.
func ValidateReply(data string) error {
	if data == errReply {
		return nil
	}

	if len(data) < minDataLen {
		return fmt.Errorf("data len: %d, err: %w", len(data), ErrDataLack)
	}

	if !strings.Contains(strings.ToLower(data), mandatoryTag) {
		return fmt.Errorf("data: %s, err: %w", data, ErrNoHeadTag)
	}

	return nil
}
