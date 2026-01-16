package redis

import (
	"CloudStorageProject-FileServer/pkg/models"
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Redis struct {
	pool *redis.Client
}

func NewRedis() (*Redis, error) {
	rdsHost := tools.GetEnv("REDIS_HOST", "localhost")
	rdsPort := tools.GetEnvAsInt("REDIS_PORT", 6379)
	rdsPassword := tools.GetEnv("REDIS_PASSWORD", "")
	rdsDB := tools.GetEnvAsInt("REDIS_DB", 0)
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", rdsHost, rdsPort),
		Password: rdsPassword,
		DB:       rdsDB,
	})
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &Redis{
		pool: client,
	}, nil
}

func (rds *Redis) Ping() error {
	return rds.pool.Ping(context.Background()).Err()
}

func (rds *Redis) Close() {
	rds.pool.Close()
}

func (rds *Redis) SetAPIField(apiData *models.APIPGS) {
	ctx := context.Background()
	rds.pool.HSet(ctx, "apikey:"+apiData.KeyName, map[string]interface{}{
		"id":          apiData.Id,
		"name":        apiData.KeyName,
		"email":       apiData.Email,
		"createdAt":   apiData.CreatedAt,
		"lastLogin":   apiData.LastLogin,
		"cloudAccess": apiData.CloudAccess,
	})
}

func (rds *Redis) GetAPIField(apikey string) (*models.APIPGS, error) {
	ctx := context.Background()
	user, err := rds.pool.HGetAll(ctx, fmt.Sprintf("apikey:%s", apikey)).Result()
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(user["id"])
	cloudAccess := user["cloudAccess"]
	email := user["email"]
	CreatedAt, _ := time.Parse("2006-01-02 15:04:05", user["createdAt"])
	LastLogin, _ := time.Parse("2006-01-02 15:04:05", user["lastLogin"])
	return &models.APIPGS{
		Id:          id,
		KeyName:     apikey,
		Email:       email,
		CloudAccess: cloudAccess,
		CreatedAt:   CreatedAt,
		LastLogin:   LastLogin,
	}, nil
}

func (rds *Redis) DelAPIField(apikey string) {
	ctx := context.Background()
	rds.pool.HDel(ctx, fmt.Sprintf("apikey:%s", apikey))
}

func (rds *Redis) ExistsAPIField(apikey string) bool {
	ctx := context.Background()
	exist, err := rds.pool.Exists(ctx, fmt.Sprintf("apikey:%s", apikey)).Result()
	if err != nil {
		//FatalError
		return false
	}
	if exist == 0 {
		return false
	}
	return true
}
