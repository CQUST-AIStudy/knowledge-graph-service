package storage

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"knowledgegraph/internal/config"
)

// CacheClient 封装 Redis 缓存操作。
// 当 Redis 不可用或未启用时，所有方法安全降级（返回 nil/false，不阻断服务）。
type CacheClient struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisClient 创建 Redis 客户端。enabled 为 false 时返回 nil 客户端（降级模式）。
func NewRedisClient(cfg config.RedisConfig, cacheCfg config.CacheConfig) (*CacheClient, error) {
	if !cfg.Enabled {
		return &CacheClient{client: nil, ttl: cacheCfg.TTL}, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return &CacheClient{client: nil, ttl: cacheCfg.TTL}, nil // 降级：连接失败不致命
	}

	return &CacheClient{client: client, ttl: cacheCfg.TTL}, nil
}

// Close 关闭 Redis 连接。
func (c *CacheClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Ping 检查 Redis 连通性。
func (c *CacheClient) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return errors.New("redis: client is nil (disabled or unreachable)")
	}
	return c.client.Ping(ctx).Err()
}

// IsEnabled 返回 Redis 是否可用。
func (c *CacheClient) IsEnabled() bool {
	return c != nil && c.client != nil
}

// ---- 图谱缓存 ----

// graphKey 返回图谱缓存的键。
func graphKey(graphCode string) string {
	return "kg:graph:" + graphCode
}

// listKey 返回图谱列表缓存的键。
const listKey = "kg:graph:list"

// GetGraphJSON 从缓存获取图谱 JSON。
func (c *CacheClient) GetGraphJSON(ctx context.Context, graphCode string) (string, error) {
	if c == nil || c.client == nil {
		return "", redis.Nil
	}
	return c.client.Get(ctx, graphKey(graphCode)).Result()
}

// SetGraphJSON 将图谱 JSON 写入缓存。
func (c *CacheClient) SetGraphJSON(ctx context.Context, graphCode string, data []byte) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, graphKey(graphCode), data, c.ttl).Err()
}

// DelGraph 删除图谱缓存。
func (c *CacheClient) DelGraph(ctx context.Context, graphCode string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(ctx, graphKey(graphCode)).Err()
}

// GetListJSON 从缓存获取图谱列表 JSON。
func (c *CacheClient) GetListJSON(ctx context.Context) (string, error) {
	if c == nil || c.client == nil {
		return "", redis.Nil
	}
	return c.client.Get(ctx, listKey).Result()
}

// SetListJSON 将图谱列表 JSON 写入缓存。
func (c *CacheClient) SetListJSON(ctx context.Context, data []byte) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, listKey, data, c.ttl).Err()
}

// DelList 删除图谱列表缓存。
func (c *CacheClient) DelList(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(ctx, listKey).Err()
}

// DelAll 删除所有知识图谱相关缓存。
func (c *CacheClient) DelAll(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	_ = c.DelList(ctx)
	// 使用 SCAN 模式删除所有 kg:graph:* 键
	iter := c.client.Scan(ctx, 0, "kg:graph:*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}
