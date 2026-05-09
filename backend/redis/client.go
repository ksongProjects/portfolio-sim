package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(host string, port int, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

func (c *Client) XAdd(ctx context.Context, stream string, values ...interface{}) error {
	return c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

func (c *Client) XRead(ctx context.Context, streams []string, count int64) ([]redis.XStream, error) {
	return c.rdb.XRead(ctx, &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   0,
	}).Result()
}

func (c *Client) XReadGroup(ctx context.Context, group, consumer, stream string, count int64) ([]redis.XStream, error) {
	return c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    0,
	}).Result()
}

func (c *Client) XLen(ctx context.Context, stream string) (int64, error) {
	return c.rdb.XLen(ctx, stream).Result()
}

func (c *Client) CreateGroup(ctx context.Context, stream string, group string) error {
	return c.rdb.XGroupCreate(ctx, stream, group, "0").Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}