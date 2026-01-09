package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	pool  *redis.Client
	ctxDB context.Context
	clsDB context.CancelFunc
}
