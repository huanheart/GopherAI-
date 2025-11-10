package redis

import (
	"GopherAI/config"
	"context"
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



func CheckCaptchaForEmail(email, userInput string) (bool, error) {
	key := GenerateCaptcha(email)


	storedCaptcha, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {

			return false, nil
		}

		return false, err
	}



	if strings.EqualFold(storedCaptcha, userInput) {

		// 验证成功后删除 key
		if err := Rdb.Del(ctx, key).Err(); err != nil {

		} else {

		}
		return true, nil
	}


	return false, nil
}
