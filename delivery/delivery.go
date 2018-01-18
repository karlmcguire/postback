package main

import (
	"os"

	"github.com/go-redis/redis"
)

func main() {
	var (
		db = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
		})
		postback []string
		err      error
	)

	println("watching redis at " + os.Getenv("REDIS_ADDR"))

	for {
		if postback, err = db.BLPop(0, "postbacks").Result(); err != nil {
			// TODO: handle errors better
			panic(err)
		}

		println(postback[1])

		//go NewPostback(db, postback[1])
		NewPostback(db, postback[1])
	}
}
