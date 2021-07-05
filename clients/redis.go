package client

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type Client struct {
	Redis   *redis.Client
	Context context.Context
}

// Create a new Redis Client to be consumed
func NewClient(ctx context.Context, host, password string) *Client {
	redis := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})

	return &Client{
		Redis:   redis,
		Context: ctx,
	}
}

// HealthCheck checks to see if the current connection to Redis is valid
func (c *Client) HealthCheck() error {
	res, err := c.Redis.Ping(c.Context).Result()
	if err != nil || res != "PONG" {
		return errors.Wrap(err, "Cannot connect to Redis server\n")
	}
	return nil
}

// DelHashKey deletes existing items from a specific key in Redis
func (c *Client) DelHashKey(key string, fields ...string) error {
	err := c.Redis.HDel(c.Context, key, fields...).Err()
	if err != nil {
		return errors.New("Failed to delete existing hash. Redis could be empty\n")
	}
	return nil
}

// HashSet sets a key for a specific hash in Redis
func (c *Client) HashSet(key string, value map[string]interface{}) error {
	err := c.Redis.HSet(c.Context, key, value).Err()
	if err != nil {
		return errors.Wrapf(err, "Failed to add %s to Redis.\n", key)
	}
	return nil
}

// RightPush pushes a value to the right of a key in Redis
func (c *Client) RightPush(key, value string) error {
	err := c.Redis.RPush(c.Context, key, value).Err()
	if err != nil {
		return errors.Wrapf(err, "Failed to add %s to %s\n", value, key)
	}
	return nil
}

// ListLength lists the length of a specific key in Redis
func (c *Client) ListLength(key string) (int64, error) {
	length, err := c.Redis.LLen(c.Context, key).Result()
	if err != nil {
		return 0, errors.Wrapf(err, "Could not read length of %s\n", key)
	}
	return length, nil
}

// LeftPop pops a specific key from Redis
func (c *Client) LeftPop(key string) (string, error) {
	workItemId, err := c.Redis.LPop(c.Context, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", err
		} else {
			return "", errors.Wrapf(err, "%s could not retrieve value\n", key)
		}
	}
	return workItemId, nil
}

// HashGetAll retrieves all of the values from a specific key in Redis
func (c *Client) HashGetAll(key string) (map[string]string, error) {
	workItem, err := c.Redis.HGetAll(c.Context, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err
		} else {
			return nil, errors.Wrapf(err, "%v could not retrieve value\n", workItem)
		}
	}
	return workItem, nil
}

// Publish publishes to a channel in Redis
func (c *Client) Publish(completedChan, key string, workerID int) error {
	err := c.Redis.Publish(c.Context, completedChan, key).Err()
	if err != nil {
		return errors.Wrapf(err, "%d could not publish %s to %s\n", workerID, key, completedChan)
	}
	return nil
}

// Subscribe subscribes to a specific channel in Redis
func (c *Client) Subscribe(completedChan string) *redis.PubSub {
	return c.Redis.Subscribe(c.Context, completedChan)
}
