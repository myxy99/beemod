package redis

import (
	"fmt"
	"github.com/beego-dev/beemod/pkg/cache/redis"
	"github.com/beego-dev/beemod/pkg/logger"
	"github.com/beego-dev/beemod/pkg/token/standard"
)

var tokenKeyPattern string

func InitTokenKeyPattern(pattern string) {
	tokenKeyPattern = pattern
}

// 如果你希望使用这个实现来作为token的实现，那么需要在配置文件里面设置：
// [muses.logger.system]
//    ...logger的配置
// [muses.redis.default]
//    ...mysql的配置
// [muses.token.jwt.redis]
//    logger = "system"
//    client = "default"
// 而后将Register()方法注册进去muses.Container(...)中
type redisTokenAccessor struct {
	standard.JwtTokenAccessor
	logger *logger.Client
	client *redis.Client
}

func InitRedisTokenAccessor(logger *logger.Client, client *redis.Client) standard.TokenAccessor {
	return &redisTokenAccessor{
		JwtTokenAccessor: standard.JwtTokenAccessor{},
		logger:           logger,
		client:           client,
	}
}

func (accessor *redisTokenAccessor) CreateAccessToken(uid int, startTime int64) (resp standard.AccessTokenTicket, err error) {

	// using the uid as the jwtId
	tokenString, err := accessor.EncodeAccessToken(uid, uid, startTime)
	if err != nil {
		return
	}

	_, err = accessor.client.Set(fmt.Sprintf(tokenKeyPattern, uid), tokenString, int(standard.AccessTokenExpireInterval))
	if err != nil {
		return
	}
	resp.AccessToken = tokenString
	resp.ExpiresIn = standard.AccessTokenExpireInterval
	return
}

func (accessor *redisTokenAccessor) CheckAccessToken(tokenStr string) bool {
	sc, err := accessor.DecodeAccessToken(tokenStr)
	if err != nil {
		return false
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	info, err := accessor.client.Get(fmt.Sprintf(tokenKeyPattern, uidInt))
	if err != nil {
		return false
	}
	// info 为nil，说明数据不存在
	if info == nil {
		return false
	}
	return true
}

func (accessor *redisTokenAccessor) RefreshAccessToken(tokenStr string, startTime int64) (resp standard.AccessTokenTicket, err error) {
	sc, err := accessor.DecodeAccessToken(tokenStr)
	if err != nil {
		return
	}
	uid := sc["jti"].(float64)
	uidInt := int(uid)
	return accessor.CreateAccessToken(uidInt, startTime)
}
