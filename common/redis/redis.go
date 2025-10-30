package redis

import (
	"GopherAI/config"
	"context"
	"log"
	"strconv"
	"strings"
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

// func CheckCaptchaForEmail(email, userInput string) (bool, error) {
// 	key := GenerateCaptcha(email)

// 	// 从 Redis 获取验证码
// 	storedCaptcha, err := Rdb.Get(ctx, key).Result()
// 	if err != nil {
// 		if err == redis.Nil {
// 			// Redis 中没有这个 key，说明验证码不存在或已过期
// 			return false, nil
// 		}
// 		return false, err // 其他 Redis 错误
// 	}

// 	// 比较验证码是否一致
// 	return storedCaptcha == userInput, nil
// }

func CheckCaptchaForEmail(email, userInput string) (bool, error) {
	key := GenerateCaptcha(email)
	log.Printf("Checking captcha for email: %s, Redis key: %s, user input: %s", email, key, userInput)

	storedCaptcha, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Captcha expired or not found for key: %s", key)
			return false, nil
		}
		log.Printf("Redis error when checking captcha for key %s: %v", key, err)
		return false, err
	}

	log.Printf("Stored captcha from Redis: %s", storedCaptcha)

	if strings.EqualFold(storedCaptcha, userInput) {
		log.Printf("Captcha match successful for email: %s", email)
		// 验证成功后删除 key
		if err := Rdb.Del(ctx, key).Err(); err != nil {
			log.Printf("Failed to delete captcha key %s: %v", key, err)
		} else {
			log.Printf("Deleted captcha key: %s", key)
		}
		return true, nil
	}

	log.Printf("Captcha mismatch for email: %s", email)
	return false, nil
}
