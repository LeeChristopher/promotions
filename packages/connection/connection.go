package connection

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

var (
	Db      *gorm.DB
	Redis   *redis.Client
	Limiter *redis.Client
)

func GetDbConnect() (db *gorm.DB, err error) {
	dbConStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", beego.AppConfig.String("db::username"), beego.AppConfig.String("db::password"), beego.AppConfig.String("db::host"), beego.AppConfig.String("db::database"))
	db, err = gorm.Open("mysql", dbConStr)
	if err == nil {
		db.SingularTable(true)
		db.DB().SetMaxIdleConns(20)
		db.DB().SetMaxOpenConns(1000)
		db.DB().SetConnMaxLifetime(2 * time.Minute)
		Db = db
	}

	return db, err
}

//连接redis
func GetRedisClient() (client *redis.Client, err error) {
	db, err := beego.AppConfig.Int("redis::database")
	if err != nil {
		return nil, err
	}
	limiterDb, err := beego.AppConfig.Int("limiter::database")
	if err != nil {
		return nil, err
	}
	//Redis连接
	redisOption := &redis.Options{
		Addr: beego.AppConfig.String("redis::host"),
		DB:   db,
	}
	if len(beego.AppConfig.String("redis::password")) > 0 {
		redisOption.Password = beego.AppConfig.String("redis::password")
	}
	client = redis.NewClient(redisOption)
	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}
	//限流实例
	redisOption.DB = limiterDb
	limiterClient := redis.NewClient(redisOption)
	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}
	Redis = client
	Limiter = limiterClient

	return client, nil
}
