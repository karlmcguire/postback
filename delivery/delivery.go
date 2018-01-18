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

	// listen forever
	for {
		// block until a new postback:[uuid] value is pushed to the postbacks
		// list
		if postback, err = db.BLPop(0, "postbacks").Result(); err != nil {
			// TODO: handle errors better
			panic(err)
		}

		// since redis operations are atomic, we can be sure there's only one
		// goroutine with each postback:[uuid] key, so now just create a new
		// postback object to do the work
		go NewPostback(db, postback[1])
	}
}
