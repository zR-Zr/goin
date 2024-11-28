package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/zR-Zr/goin/pkg/config"
)

const (
	ErrInvalidToken        = "invalid token"        // Token 无效
	ErrTokenExpired        = "token timeout"        // token 过期
	ErrTokenGenerateFailed = "token genrate failed" // token生成失败
	ErrAuthInfoNotFound    = "auth info not found"  // 认证信息未找到
)

const (
	TokenTypeRefresh = "refresh"
	TokenTypeAccess  = "access"
)

type JWTUser struct {
	ID          uint   `josn:"id"`
	Username    string `json:"username"`
	Type        string `json:"type"`
	IsAnonymous bool   `josn:"-"` // 是否是匿名用户
}

// CustomClaims 自定义 JWT 声明
type CustomClaims struct {
	ID                   uint   `json:"id"`
	Username             string `json:"username"`
	Type                 string `json:"type"` // 用户类型,例如 admin,  user
	jwt.RegisteredClaims        // 内置声明
}

func NewCustomClaims(id uint, username, userType string, duration time.Duration, tokenType string) CustomClaims {
	now := time.Now()

	return CustomClaims{
		id,
		username,
		userType,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(now),               // 签发时间
			NotBefore: jwt.NewNumericDate(now),               // 生效时间
			Issuer:    "admin",                               // 签发者
			Subject:   tokenType,                             // token类型
			ID:        uuid.NewString(),                      // jwt id,
		},
	}
}

func (cc CustomClaims) Valid() error {
	now := time.Now()

	// 验证生效时间
	if cc.NotBefore != nil && now.Before(cc.NotBefore.Time) {
		return fmt.Errorf("token is not yet valid")
	}

	// 验证过期时间
	if cc.ExpiresAt != nil && now.After(cc.ExpiresAt.Time) {
		return errors.New(ErrTokenExpired) // fmt.Errorf("token is expired")
	}

	return nil
}

// ----------------- jwt 认证器 ---------------------

// JWTAuthenticator JWT 认证器
type JWTAuthenticator struct {
	Secret          []byte        // 密钥
	accessTokenExp  time.Duration // access token 过期时间
	refreshTokenExp time.Duration
	rdb             *redis.Client
}

// NewJWTAuth 创建JWT认证器
func NewJWTAuth(cfg *config.Config) *JWTAuthenticator {

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	au := &JWTAuthenticator{
		Secret:          []byte(cfg.JWT.Secret), // []byte(viper.GetString("jwt.secret")),
		accessTokenExp:  cfg.JWT.AccessTokenExp, // viper.GetDuration("jwt.access_token_expire"),
		refreshTokenExp: cfg.JWT.RefreshTokenExp,
		rdb:             rdb,
	}

	log.Print("au :", au)

	return au
}

// GenrateToken 生成 Token
func (ja *JWTAuthenticator) GenrateToken(user *JWTUser) (string, error) {
	claims := NewCustomClaims(user.ID, user.Username, user.Type, ja.accessTokenExp, TokenTypeAccess)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 是用 HS256 算法进行签名
	tokenString, err := token.SignedString(ja.Secret)          // 使用密钥进行签名
	if err != nil {
		return "", errors.New(ErrTokenGenerateFailed) // fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成 refresh token
func (ja *JWTAuthenticator) GenerateRefreshToken(user *JWTUser) (string, error) {
	claims := NewCustomClaims(user.ID, "", "", viper.GetDuration("jwt.refresh_token_expire"), TokenTypeRefresh)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 是用 HS256 算法进行签名
	tokenString, err := token.SignedString(ja.Secret)          // 使用密钥进行签名
	if err != nil {
		return "", errors.New(ErrTokenGenerateFailed) //fmt.Errorf("failed to sign  refresh token: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析 token
func (ja *JWTAuthenticator) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 校验 signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(ErrInvalidToken) // fmt.Errorf("unexpedcted signing method: %v", t.Header["alg"])
		}

		return ja.Secret, nil
	})
	if err != nil {
		return nil, errors.New(ErrInvalidToken) // fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid { // 校验token 和声明是否有效
		// 检查 token 是否在黑名单中
		blacklisted, err := ja.IsTokenInBlacklisted(tokenString)
		if err != nil {
			return nil, fmt.Errorf("failed to check token blacklist: %w", err)
		}
		if blacklisted {
			return nil, errors.New(ErrInvalidToken) // errors.New("token is blacklisted")
		}

		return claims, nil
	}

	return nil, errors.New(ErrInvalidToken) //errors.New("invalid token")
}

// RefreshToken 刷新Token
func (ja *JWTAuthenticator) RefreshToken(refreshTokenString string) (map[string]interface{}, error) {
	// 1. 解析 refresh token
	claims, err := ja.ParseToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// 2. 检查 refresh token 是否有效
	if err := claims.Valid(); err != nil {
		return nil, err
	}

	// 3. 检查 refresh token 是否在黑名单中
	// TODO: 实现黑名单逻辑
	blacklisted, err := ja.IsTokenInBlacklisted(refreshTokenString) // 检查黑名单
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, errors.New("refresh token is blacklisted")
	}

	if claims.Subject != TokenTypeRefresh {
		return nil, errors.New("token type error")
	}

	// 4. 生成新的 access token 和 refresh token
	accessToken, err := ja.GenrateToken(&JWTUser{
		ID:       claims.ID,
		Username: claims.Username,
		Type:     claims.Type,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "生成token失败")
	}

	refreshToken, err := ja.GenerateRefreshToken(&JWTUser{
		ID:       claims.ID,
		Username: claims.Username,
		Type:     claims.Type,
	})

	if err != nil {
		return nil, errors.WithMessage(err, "生成refresh token 失败")
	}

	return map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

// IsTokenInBlacklisted 检查 token 是否在黑名单中
func (ja *JWTAuthenticator) IsTokenInBlacklisted(token string) (bool, error) {
	// 使用 Redis 的 EXISTS 命令检查 token 是否存在
	exists, err := ja.rdb.Exists(context.Background(), token).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists == 1, nil
}

// AddTokenToBlacklist 将 token 添加到黑名单
func (ja *JWTAuthenticator) AddTokenToBlacklist(token string, expiresAt time.Time) error {
	// 使用 Redis 的 SETEX 命令添加 token 到黑名单,并设置过期时间
	err := ja.rdb.SetEX(context.Background(), token, "", time.Until(expiresAt)).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// Valid 验证 Claims 是否有效
func Valid(claims *jwt.RegisteredClaims) error {
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return errors.New("refresh token is expired")
	}

	return nil
}
