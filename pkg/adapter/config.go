package adapter

import (
	"context"
)

type Config struct{}

func (c *Config) Validate(_ context.Context) error {
	return nil
}
