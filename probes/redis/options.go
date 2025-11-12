package redis

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Option defines a configuration option for the Redis Probe.
type Option func(*options) error

// SetCheck defines a check that attempts to set a key-value pair
// in Redis with an optional expiration.
//
//   - Key: The key name that will be attempted to be set.
//   - Value: The value to set for the specified key (Optional).
//   - Expiration: The duration after which the key should expire.
//     A zero value means the key will not expire (Optional).
type SetCheck struct {
	Key        string
	Value      string
	Expiration time.Duration
}

// GetCheck defines a check that attempts to get a key from Redis
// and (optionally) verify its value.
//
//   - Key: The key name that will be attempted to be retrieved.
//   - ExpectedValue: The expected value for the specified key (Optional).
//   - NoErrorOnMissingKey: If set to true, the check will not return
//     an error if the key is missing in Redis.
type GetCheck struct {
	Key                 string
	ExpectedValue       string
	NoErrorOnMissingKey bool
}

type options struct {
	set *SetCheck
	get *GetCheck

	dsn       string
	redisOpts *redis.Options
}

// WithSetChecker configures a SetCheck to be performed during the health check.
func WithSetChecker(c SetCheck) Option {
	return func(o *options) error {
		if c.Key == "" {
			return fmt.Errorf("set check key cannot be empty")
		}

		o.set = &c

		return nil
	}
}

// WithGetChecker configures a GetCheck to be performed during the health check.
func WithGetChecker(c GetCheck) Option {
	return func(o *options) error {
		if c.Key == "" {
			return fmt.Errorf("get check key cannot be empty")
		}

		o.get = &c

		return nil
	}
}
