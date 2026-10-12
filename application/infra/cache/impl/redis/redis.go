package redis

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"gosilite/application/infra/cache"
)

// New creates a new Redis cache client using environment variables
// 创建一个新的Redis缓存客户端，使用环境变量配置
func New() cache.CacheClient {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")

	return NewWithConfig(addr, password)
}

// NewWithConfig creates a new Redis cache client with the given address and password
// 使用给定的地址和密码创建一个新的Redis缓存客户端
func NewWithConfig(addr, password string) cache.CacheClient {
	// Set the default empty value error to redis.Nil
	// 将默认空值错误设置为redis.Nil
	cache.SetEmptyValueError(redis.Nil)

	// Create a new Redis client with connection pool configuration
	// 创建一个新的Redis客户端，配置连接池
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // Redis server address
		Password: password, // Redis password (empty for no password)
		DB:       0,        // Default database index

		// Connection pool settings
		// 连接池设置
		PoolSize:        100,             // Maximum number of connections
		MinIdleConns:    10,              // Minimum number of idle connections
		MaxIdleConns:    30,              // Maximum number of idle connections
		ConnMaxIdleTime: 5 * time.Minute, // Idle connection timeout

		// Timeout settings
		// 超时设置
		DialTimeout:  5 * time.Second, // Connection establishment timeout
		ReadTimeout:  3 * time.Second, // Read operation timeout
		WriteTimeout: 3 * time.Second, // Write operation timeout
	})

	return &redisClient{client: rdb}
}

// redisClient implements the CacheClient interface using Redis
// redisClient使用Redis实现CacheClient接口

type redisClient struct {
	client *redis.Client
}

// CreatePipeline creates a new pipeline for batch operations
// 创建一个新的管道用于批量操作
func (r *redisClient) CreatePipeline() cache.Pipeline {
	p := r.client.Pipeline()
	return &redisPipeline{pipeline: p}
}

// Delete removes one or more keys from cache
// 从缓存中删除一个或多个键
func (r *redisClient) Delete(ctx context.Context, keys ...string) cache.IntCmd {
	return r.client.Del(ctx, keys...)
}

// Exists checks if one or more keys exist in cache
// 检查一个或多个键是否存在于缓存中
func (r *redisClient) Exists(ctx context.Context, keys ...string) cache.IntCmd {
	return r.client.Exists(ctx, keys...)
}

// Expire sets expiration time for a key
// 为键设置过期时间
func (r *redisClient) Expire(ctx context.Context, key string, expiration time.Duration) cache.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

// Get retrieves a value by key from cache
// 根据键从缓存中获取值
func (r *redisClient) Get(ctx context.Context, key string) cache.StringCmd {
	return r.client.Get(ctx, key)
}

// HGetAll retrieves all field-value pairs from a hash
// 从哈希中获取所有字段值对
func (r *redisClient) HGetAll(ctx context.Context, key string) cache.MapStringStringCmd {
	return r.client.HGetAll(ctx, key)
}

// HSet sets multiple field-value pairs to a hash
// 向哈希中设置多个字段值对
func (r *redisClient) HSet(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

// HDel deletes one or more fields from a hash
// 删除哈希中的一个或多个字段
func (r *redisClient) HDel(ctx context.Context, key string, fields ...string) cache.IntCmd {
	return r.client.HDel(ctx, key, fields...)
}

// Incr increments a key by 1
// 将键的值增加1
func (r *redisClient) Incr(ctx context.Context, key string) cache.IntCmd {
	return r.client.Incr(ctx, key)
}

// IncrBy increments a key by the given value
// 将键的值增加指定数量
func (r *redisClient) IncrBy(ctx context.Context, key string, value int64) cache.IntCmd {
	return r.client.IncrBy(ctx, key, value)
}

// LIndex retrieves an element from a list by its index
// 根据索引从列表中获取元素
func (r *redisClient) LIndex(ctx context.Context, key string, index int64) cache.StringCmd {
	return r.client.LIndex(ctx, key, index)
}

// LPop pops a value from the left of a list
// 从列表左侧弹出一个值
func (r *redisClient) LPop(ctx context.Context, key string) cache.StringCmd {
	return r.client.LPop(ctx, key)
}

// LPush pushes one or more values to the left of a list
// 向列表左侧推入一个或多个值
func (r *redisClient) LPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return r.client.LPush(ctx, key, values...)
}

// LRange retrieves elements from a list within a range
// 从列表中获取指定范围的元素
func (r *redisClient) LRange(ctx context.Context, key string, start int64, stop int64) cache.StringSliceCmd {
	return r.client.LRange(ctx, key, start, stop)
}

// LSet sets the value of an element in a list by its index
// 根据索引设置列表中元素的值
func (r *redisClient) LSet(ctx context.Context, key string, index int64, value interface{}) cache.StatusCmd {
	return r.client.LSet(ctx, key, index, value)
}

// RPush pushes one or more values to the right of a list
// 向列表右侧推入一个或多个值
func (r *redisClient) RPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return r.client.RPush(ctx, key, values...)
}

// Set stores a key-value pair with expiration
// 存储键值对并设置过期时间
func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) cache.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

const renewLeaseScript = `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('pexpire', KEYS[1], ARGV[2]) else return 0 end`
const releaseLeaseScript = `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end`

func (r *redisClient) AcquireLease(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, token, ttl).Result()
}

func (r *redisClient) RenewLease(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	n, err := r.client.Eval(ctx, renewLeaseScript, []string{key}, token, ttl.Milliseconds()).Int64()
	return n == 1, err
}

func (r *redisClient) ReleaseLease(ctx context.Context, key, token string) (bool, error) {
	n, err := r.client.Eval(ctx, releaseLeaseScript, []string{key}, token).Int64()
	return n == 1, err
}

// ScanAll 有界 SCAN 匹配 key（禁止 KEYS）。
func (r *redisClient) ScanAll(ctx context.Context, match string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}
	var (
		out    []string
		cursor uint64
	)
	for {
		keys, next, err := r.client.Scan(ctx, cursor, match, 200).Result()
		if err != nil {
			return out, err
		}
		out = append(out, keys...)
		if len(out) >= limit {
			return out[:limit], nil
		}
		if next == 0 {
			return out, nil
		}
		cursor = next
	}
}

// TypeOf 返回 key 的 Redis 类型。
func (r *redisClient) TypeOf(ctx context.Context, key string) (string, error) {
	return r.client.Type(ctx, key).Result()
}

// redisPipeline implements the Pipeline interface for Redis
// redisPipeline为Redis实现Pipeline接口

type redisPipeline struct {
	pipeline redis.Pipeliner
}

// Delete removes one or more keys from cache (pipeline operation)
// 从缓存中删除一个或多个键（管道操作）
func (p *redisPipeline) Delete(ctx context.Context, keys ...string) cache.IntCmd {
	return p.pipeline.Del(ctx, keys...)
}

// Execute executes all commands in the pipeline
// 执行管道中的所有命令
func (p *redisPipeline) Execute(ctx context.Context) ([]cache.Cmder, error) {
	cmders, err := p.pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return convertCmds(cmders), nil
}

// convertCmds converts redis.Cmder slice to cache.Cmder slice
// 将redis.Cmder切片转换为cache.Cmder切片
func convertCmds(cmders []redis.Cmder) []cache.Cmder {
	result := make([]cache.Cmder, 0, len(cmders))
	for _, cmder := range cmders {
		result = append(result, cmder)
	}
	return result
}

// Exists checks if one or more keys exist in cache (pipeline operation)
// 检查一个或多个键是否存在于缓存中（管道操作）
func (p *redisPipeline) Exists(ctx context.Context, keys ...string) cache.IntCmd {
	return p.pipeline.Exists(ctx, keys...)
}

// Expire sets expiration time for a key (pipeline operation)
// 为键设置过期时间（管道操作）
func (p *redisPipeline) Expire(ctx context.Context, key string, expiration time.Duration) cache.BoolCmd {
	return p.pipeline.Expire(ctx, key, expiration)
}

// Get retrieves a value by key from cache (pipeline operation)
// 根据键从缓存中获取值（管道操作）
func (p *redisPipeline) Get(ctx context.Context, key string) cache.StringCmd {
	return p.pipeline.Get(ctx, key)
}

// HGetAll retrieves all field-value pairs from a hash (pipeline operation)
// 从哈希中获取所有字段值对（管道操作）
func (p *redisPipeline) HGetAll(ctx context.Context, key string) cache.MapStringStringCmd {
	return p.pipeline.HGetAll(ctx, key)
}

// HSet sets multiple field-value pairs to a hash (pipeline operation)
// 向哈希中设置多个字段值对（管道操作）
func (p *redisPipeline) HSet(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return p.pipeline.HSet(ctx, key, values...)
}

// HDel deletes one or more fields from a hash (pipeline operation)
// 删除哈希中的一个或多个字段（管道操作）
func (p *redisPipeline) HDel(ctx context.Context, key string, fields ...string) cache.IntCmd {
	return p.pipeline.HDel(ctx, key, fields...)
}

// Incr increments a key by 1 (pipeline operation)
// 将键的值增加1（管道操作）
func (p *redisPipeline) Incr(ctx context.Context, key string) cache.IntCmd {
	return p.pipeline.Incr(ctx, key)
}

// IncrBy increments a key by the given value (pipeline operation)
// 将键的值增加指定数量（管道操作）
func (p *redisPipeline) IncrBy(ctx context.Context, key string, value int64) cache.IntCmd {
	return p.pipeline.IncrBy(ctx, key, value)
}

// LIndex retrieves an element from a list by its index (pipeline operation)
// 根据索引从列表中获取元素（管道操作）
func (p *redisPipeline) LIndex(ctx context.Context, key string, index int64) cache.StringCmd {
	return p.pipeline.LIndex(ctx, key, index)
}

// LPop pops a value from the left of a list (pipeline operation)
// 从列表左侧弹出一个值（管道操作）
func (p *redisPipeline) LPop(ctx context.Context, key string) cache.StringCmd {
	return p.pipeline.LPop(ctx, key)
}

// LPush pushes one or more values to the left of a list (pipeline operation)
// 向列表左侧推入一个或多个值（管道操作）
func (p *redisPipeline) LPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return p.pipeline.LPush(ctx, key, values...)
}

// LRange retrieves elements from a list within a range (pipeline operation)
// 从列表中获取指定范围的元素（管道操作）
func (p *redisPipeline) LRange(ctx context.Context, key string, start int64, stop int64) cache.StringSliceCmd {
	return p.pipeline.LRange(ctx, key, start, stop)
}

// LSet sets the value of an element in a list by its index (pipeline operation)
// 根据索引设置列表中元素的值（管道操作）
func (p *redisPipeline) LSet(ctx context.Context, key string, index int64, value interface{}) cache.StatusCmd {
	return p.pipeline.LSet(ctx, key, index, value)
}

// CreatePipeline returns the current pipeline (for pipeline nesting)
// 返回当前管道（用于管道嵌套）
func (p *redisPipeline) CreatePipeline() cache.Pipeline {
	return p
}

// RPush pushes one or more values to the right of a list (pipeline operation)
// 向列表右侧推入一个或多个值（管道操作）
func (p *redisPipeline) RPush(ctx context.Context, key string, values ...interface{}) cache.IntCmd {
	return p.pipeline.RPush(ctx, key, values...)
}

// Set stores a key-value pair with expiration (pipeline operation)
// 存储键值对并设置过期时间（管道操作）
func (p *redisPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) cache.StatusCmd {
	return p.pipeline.Set(ctx, key, value, expiration)
}

// ScanAll 管道内不支持 SCAN。
func (p *redisPipeline) ScanAll(ctx context.Context, match string, limit int) ([]string, error) {
	return nil, errors.New("cache: ScanAll not supported in pipeline")
}

// TypeOf 管道内不支持 TYPE。
func (p *redisPipeline) TypeOf(ctx context.Context, key string) (string, error) {
	return "", errors.New("cache: TypeOf not supported in pipeline")
}
