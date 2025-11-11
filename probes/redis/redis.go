package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/botchris/go-health"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

const pongResponse = "PONG"

type redisProbe struct {
	opts *options
}

// New creates new Redis health check that verifies that a connection to the Redis server
// can be established and a PING command returns the expected PONG response.
func New(dsn string, o ...Option) (health.Probe, error) {
	redisOptions, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse redis dsn", err)
	}

	if redisOptions.MaintNotificationsConfig == nil {
		redisOptions.MaintNotificationsConfig = &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		}
	}

	opts := &options{
		dsn:       dsn,
		redisOpts: redisOptions,
	}

	for i := range o {
		if oErr := o[i](opts); oErr != nil {
			return nil, oErr
		}
	}

	return redisProbe{opts: opts}, nil
}

func (r redisProbe) Check(ctx context.Context) (checkErr error) {
	rdb := redis.NewClient(r.opts.redisOpts)

	defer func() {
		if cErr := rdb.Close(); cErr != nil && checkErr == nil {
			checkErr = fmt.Errorf("failed to close redis client: %w", cErr)
		}
	}()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		checkErr = fmt.Errorf("%w: redis ping failed", err)

		return
	}

	if pong != pongResponse {
		checkErr = fmt.Errorf("unexpected response for redis ping: %q", pong)

		return
	}

	if checkErr = r.setChecker(ctx, rdb); checkErr != nil {
		return
	}

	checkErr = r.getChecker(ctx, rdb)

	return
}

func (r redisProbe) setChecker(ctx context.Context, rdb *redis.Client) error {
	if r.opts.set == nil {
		return nil
	}

	if err := rdb.Set(ctx, r.opts.set.Key, r.opts.set.Value, r.opts.set.Expiration).Err(); err != nil {
		return fmt.Errorf("%w: failed to set key %q", err, r.opts.set.Key)
	}

	return nil
}

func (r redisProbe) getChecker(ctx context.Context, rdb *redis.Client) error {
	if r.opts.get == nil {
		return nil
	}

	val, gErr := rdb.Get(ctx, r.opts.get.Key).Result()
	if gErr != nil {
		if errors.Is(gErr, redis.Nil) {
			if !r.opts.get.NoErrorOnMissingKey {
				return fmt.Errorf("key %q does not exist", r.opts.get.Key)
			}

			return nil
		}

		return fmt.Errorf("%w: failed to get key %q", gErr, r.opts.get.Key)
	}

	if r.opts.get.ExpectedValue != "" && val != r.opts.get.ExpectedValue {
		return fmt.Errorf("unexpected value for key %q: got %q, want %q", r.opts.get.Key, val, r.opts.get.ExpectedValue)
	}

	return nil
}
