package main

import (
	"flag"
	"log"

	"github.com/go-redis/redis"
)

var (
	REDIS_ADDR = flag.String(
		"redis_addr",
		"127.0.0.1:6379",
		"address of redis server to listen for postback objects",
	)
	REDIS_PASSWORD = flag.String(
		"redis_password",
		"none",
		"password for authenticating with redis",
	)
)

func init() {
	flag.Parse()
}

func main() {
	var (
		db = redis.NewClient(&redis.Options{
			Addr:     *REDIS_ADDR,
			Password: *REDIS_PASSWORD,
		})
		postback []string
		err      error
	)

	log.Printf("watching redis at %s\n", *REDIS_ADDR)

	// listen forever
	for {
		// block until a new postback:[uuid] value is pushed to the postbacks
		// list
		if postback, err = db.BLPop(0, "postbacks").Result(); err != nil {
			log.Print(err)
			continue
		}

		// since redis operations are atomic, we can be sure there's only one
		// goroutine with each postback:[uuid] key, so now just create a new
		// postback object to do the work
		go NewPostback(db, postback[1])
	}
}
