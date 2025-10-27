package redis

import (
	"GopherAI/config"
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

var ctx = context.Background()

func Init() {
	conf := config.GetConfig()
	host := conf.RedisConfig.RedisHost
	port := conf.RedisConfig.RedisPort
	password := conf.RedisConfig.RedisPassword
	db := conf.RedisDb
	addr := host + ":" + strconv.Itoa(port)

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

}

func SetCaptchaForEmail(email, captcha string) error {
	key := GenerateCaptcha(email)
	expire := 2 * time.Minute
	return Rdb.Set(ctx, key, captcha, expire).Err()
}

func CheckCaptchaForEmail(email, userInput string) (bool, error) {
	key := GenerateCaptcha(email)

	// 从 Redis 获取验证码
	storedCaptcha, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Redis 中没有这个 key，说明验证码不存在或已过期
			return false, nil
		}
		return false, err // 其他 Redis 错误
	}

	// 比较验证码是否一致
	return storedCaptcha == userInput, nil
}
