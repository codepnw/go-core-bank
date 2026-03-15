package transferrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=transfer_redis_repo.go -destination=transfer_redis_repo_mock.go -package=transferrepository
type TransferRedisRepository interface {
	CheckIdempotencyKey(ctx context.Context, key string) (bool, error)
	DeleteIdempotencyKey(ctx context.Context, idemKey string) error
}

type transferRedisRepository struct {
	client *redis.Client
}

func NewTransferRedisRepository(client *redis.Client) TransferRedisRepository {
	return &transferRedisRepository{client: client}
}

func (r *transferRedisRepository) CheckIdempotencyKey(ctx context.Context, idemKey string) (bool, error) {
	redisKey := r.makeTransferIdemKey(idemKey)
	
	err := r.client.SetArgs(ctx, redisKey, "processing", redis.SetArgs{
		Mode: "NX",
		TTL:  24 * time.Hour,
	}).Err()
	if err != nil {
		if err == redis.Nil {
			// set failed (found redis-key)
			return false, nil
		}
		return false, fmt.Errorf("set redis failed: %w", err)
	}
	// new redis-key can transfer
	return true, nil
}

func (r *transferRedisRepository) DeleteIdempotencyKey(ctx context.Context, idemKey string) error {
	redisKey := r.makeTransferIdemKey(idemKey)
	return r.client.Del(ctx, redisKey).Err()
}

func (r *transferRedisRepository) makeTransferIdemKey(idemKey string) string {
	return fmt.Sprintf("transfer:idem:%s", idemKey)
}
