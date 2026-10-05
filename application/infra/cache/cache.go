package cache

import (
	"context"
	"time"
)

// EmptyValue represents the empty value error, typically used when key does not exist
var EmptyValue error

// SetEmptyValueError sets the default empty value error
func SetEmptyValueError(err error) {
	EmptyValue = err
}

// CacheClient defines the main cache client interface that combines all cache operations
// 缓存客户端接口，整合所有缓存操作

type CacheClient interface {
	// CreatePipeline creates a new pipeline for batch operations
	// 创建一个新的管道用于批量操作
	CreatePipeline() Pipeline

	// StringOperations returns string-specific cache operations
	// 获取字符串类型缓存操作接口
	StringOperations

	// HashOperations returns hash-specific cache operations
	// 获取哈希类型缓存操作接口
	HashOperations

	// CoreOperations returns core cache operations
	// 获取核心缓存操作接口
	CoreOperations

	// ListOperations returns list-specific cache operations
	// 获取列表类型缓存操作接口
	ListOperations
}

// StringOperations defines string-related cache operations
// 字符串类型缓存操作接口

type StringOperations interface {
	// Set stores a key-value pair with expiration
	// 存储键值对并设置过期时间
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd

	// Get retrieves a value by key
	// 根据键获取值
	Get(ctx context.Context, key string) StringCmd

	// IncrBy increments a key by the given value
	// 将键的值增加指定数量
	IncrBy(ctx context.Context, key string, value int64) IntCmd

	// Incr increments a key by 1
	// 将键的值增加1
	Incr(ctx context.Context, key string) IntCmd
}

// HashOperations defines hash-related cache operations
// 哈希类型缓存操作接口

type HashOperations interface {
	// HSet sets multiple field-value pairs to a hash
	// 向哈希中设置多个字段值对
	HSet(ctx context.Context, key string, values ...interface{}) IntCmd

	// HGetAll retrieves all field-value pairs from a hash
	// 从哈希中获取所有字段值对
	HGetAll(ctx context.Context, key string) MapStringStringCmd
}

// CoreOperations defines core cache operations
// 核心缓存操作接口

type CoreOperations interface {
	// Delete removes one or more keys
	// 删除一个或多个键
	Delete(ctx context.Context, keys ...string) IntCmd

	// Exists checks if one or more keys exist
	// 检查一个或多个键是否存在
	Exists(ctx context.Context, keys ...string) IntCmd

	// Expire sets expiration time for a key
	// 为键设置过期时间
	Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd
}

// ListOperations defines list-related cache operations
// 列表类型缓存操作接口

type ListOperations interface {
	// LIndex retrieves an element from a list by its index
	// 根据索引从列表中获取元素
	LIndex(ctx context.Context, key string, index int64) StringCmd

	// LPush pushes one or more values to the left of a list
	// 向列表左侧推入一个或多个值
	LPush(ctx context.Context, key string, values ...interface{}) IntCmd

	// RPush pushes one or more values to the right of a list
	// 向列表右侧推入一个或多个值
	RPush(ctx context.Context, key string, values ...interface{}) IntCmd

	// LSet sets the value of an element in a list by its index
	// 根据索引设置列表中元素的值
	LSet(ctx context.Context, key string, index int64, value interface{}) StatusCmd

	// LPop pops a value from the left of a list
	// 从列表左侧弹出一个值
	LPop(ctx context.Context, key string) StringCmd

	// LRange retrieves elements from a list within a range
	// 从列表中获取指定范围的元素
	LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd
}

// Pipeline defines batch operation capabilities
// 管道接口，用于批量操作

type Pipeline interface {
	// PipelineCacheClient combines all cache operations for pipeline
	// 管道缓存客户端，整合所有管道操作
	PipelineCacheClient

	// Execute executes all commands in the pipeline
	// 执行管道中的所有命令
	Execute(ctx context.Context) ([]Cmder, error)
}

// PipelineCacheClient defines cache operations that can be used in pipeline
// 管道缓存客户端接口，定义可在管道中使用的缓存操作

type PipelineCacheClient interface {
	CacheClient
}

// Cmder defines the base command interface
// 命令接口，定义命令的基本方法

type Cmder interface {
	// Err returns the error of the command, if any
	// 返回命令的错误（如果有）
	Err() error
}

// IntCmd defines integer command interface
// 整数命令接口，定义整数命令的方法

type IntCmd interface {
	Cmder
	// Result returns the integer value and error
	// 返回整数值和错误
	Result() (int64, error)
	// Val returns the integer value (ignores error)
	// 返回整数值（忽略错误）
	Val() int64
}

// MapStringStringCmd defines hash command interface
// 哈希命令接口，定义哈希命令的方法

type MapStringStringCmd interface {
	Cmder
	// Result returns the hash map and error
	// 返回哈希映射和错误
	Result() (map[string]string, error)
	// Val returns the hash map (ignores error)
	// 返回哈希映射（忽略错误）
	Val() map[string]string
}

// BoolCmd defines boolean command interface
// 布尔命令接口，定义布尔命令的方法

type BoolCmd interface {
	Cmder
	// Result returns the boolean value and error
	// 返回布尔值和错误
	Result() (bool, error)
	// Val returns the boolean value (ignores error)
	// 返回布尔值（忽略错误）
	Val() bool
}

// StatusCmd defines status command interface
// 状态命令接口，定义状态命令的方法

type StatusCmd interface {
	Cmder
	// Result returns the status string and error
	// 返回状态字符串和错误
	Result() (string, error)
	// Val returns the status string (ignores error)
	// 返回状态字符串（忽略错误）
	Val() string
}

// StringCmd defines string command interface
// 字符串命令接口，定义字符串命令的方法

type StringCmd interface {
	Cmder
	// Result returns the string value and error
	// 返回字符串值和错误
	Result() (string, error)
	// Val returns the string value (ignores error)
	// 返回字符串值（忽略错误）
	Val() string
	// Int64 returns the value as int64 and error
	// 返回int64值和错误
	Int64() (int64, error)
	// Bytes returns the value as bytes and error
	// 返回字节数组和错误
	Bytes() ([]byte, error)
}

// StringSliceCmd defines string slice command interface
// 字符串切片命令接口，定义字符串切片命令的方法

type StringSliceCmd interface {
	Cmder
	// Result returns the string slice and error
	// 返回字符串切片和错误
	Result() ([]string, error)
	// Val returns the string slice (ignores error)
	// 返回字符串切片（忽略错误）
	Val() []string
}
