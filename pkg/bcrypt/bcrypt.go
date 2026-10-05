package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

// HashAlgorithm 定义支持的哈希算法类型
type HashAlgorithm string

const (
	// AlgorithmBCrypt 表示 bcrypt 算法
	AlgorithmBCrypt HashAlgorithm = "bcrypt"
)

// PasswordHasher 密码哈希接口
type PasswordHasher interface {
	// Hash 加密密码
	Hash(password string) (string, error)
	// Check 验证密码
	Check(hashedPassword, password string) bool
	// GetAlgorithm 获取使用的哈希算法
	GetAlgorithm() HashAlgorithm
	// 保持向后兼容性的方法
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) bool
}

// HasherConfig 哈希配置
type HasherConfig struct {
	// Algorithm 哈希算法
	Algorithm HashAlgorithm
	// Cost 哈希成本因子（仅用于 bcrypt）
	Cost int
}

// DefaultConfig 默认配置
var DefaultConfig = HasherConfig{
	Algorithm: AlgorithmBCrypt,
	Cost:      bcrypt.DefaultCost,
}

// bcryptHasher bcrypt 实现
type bcryptHasher struct {
	config HasherConfig
}

// NewPasswordHasher 创建密码处理器
func NewPasswordHasher(config ...HasherConfig) PasswordHasher {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
		// 如果成本因子未设置或无效，使用默认值
		if cfg.Cost <= 0 {
			cfg.Cost = bcrypt.DefaultCost
		}
	}

	switch cfg.Algorithm {
	case AlgorithmBCrypt:
		return &bcryptHasher{config: cfg}
	default:
		// 默认使用 bcrypt
		return &bcryptHasher{config: DefaultConfig}
	}
}

// Hash 加密密码
func (h *bcryptHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.config.Cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Check 验证密码
func (h *bcryptHasher) Check(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetAlgorithm 获取使用的哈希算法
func (h *bcryptHasher) GetAlgorithm() HashAlgorithm {
	return h.config.Algorithm
}

// HashPassword 向后兼容的密码加密方法
func (h *bcryptHasher) HashPassword(password string) (string, error) {
	return h.Hash(password)
}

// CheckPassword 向后兼容的密码验证方法
func (h *bcryptHasher) CheckPassword(hashedPassword, password string) bool {
	return h.Check(hashedPassword, password)
}

// 全局便捷函数
var defaultHasher = NewPasswordHasher()

// Hash 全局密码加密函数
func Hash(password string) (string, error) {
	return defaultHasher.Hash(password)
}

// Check 全局密码验证函数
func Check(hashedPassword, password string) bool {
	return defaultHasher.Check(hashedPassword, password)
}

// HashPassword 向后兼容的全局密码加密函数
func HashPassword(password string) (string, error) {
	return defaultHasher.HashPassword(password)
}

// CheckPassword 向后兼容的全局密码验证函数
func CheckPassword(hashedPassword, password string) bool {
	return defaultHasher.CheckPassword(hashedPassword, password)
}

// SetDefaultHasher 设置默认密码处理器
func SetDefaultHasher(hasher PasswordHasher) {
	defaultHasher = hasher
}
