package main

import (
	"errors"
	"time"
)

// config holds the configuration for Gitea client
type config struct {
	Server     string
	Token      string
	SkipVerify bool
	Timeout    time.Duration
	RetryCount int
}

// Validate validates the configuration and sets defaults
func (c *config) Validate() error {
	if c.Server == "" {
		return errors.New("server URL is required")
	}
	if c.Token == "" {
		return errors.New("token is required")
	}
	// Gitea tokens are typically 40 characters (hex format)
	if len(c.Token) < 40 {
		return errors.New("token appears invalid (expected 40 characters)")
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second // default timeout
	}
	if c.RetryCount == 0 {
		c.RetryCount = 3 // default retry count
	}
	return nil
}
