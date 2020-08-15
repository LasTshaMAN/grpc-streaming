package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LasTshaMAN/streaming/internal/config"
)

func Test_configuredUsers(t *testing.T) {
	want := config.Config{
		URLs: []string{
			"https://golang.org",
			"https://www.google.com",
			"https://www.bbc.co.uk",
			"https://www.github.com",
			"https://www.gitlab.com",
			"https://www.duckduckgo.com",
			"https://www.atlasian.com",
			"https://www.twitter.com",
			"https://www.facebook.com",
		},
		MinTimeout:       10 * time.Second,
		MaxTimeout:       100 * time.Second,
		NumberOfRequests: 3,
	}

	got, err := config.Parse("../../config/config.yml")

	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
