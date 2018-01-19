package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	REQUEST_COUNT = 1
	DATA_COUNT    = 10
)

func Producer() {
	var (
		r = NewRequest(
			"GET",
			os.Getenv("TESTING_ADDR")+"/?time={time}&id={id}",
		)
		err error
	)

	// add DATA_COUNT data objects to the request with the time and id params so
	// that we can estimate performance by subtracting the difference when
	// received
	for i := 0; i < DATA_COUNT; i++ {
		data := map[string]string{
			"time": fmt.Sprintf("%d", time.Now().Unix()),
			"id":   fmt.Sprintf("%d", i),
		}

		// keep in memory so we can track from the consumer
		//Sent[data["id"]] = data["time"]

		r.AddData(data)
	}

	// send each request to INGESTION_ADDR
	for i := 0; i < REQUEST_COUNT; i++ {
		if err = r.Send(os.Getenv("INGESTION_ADDR")); err != nil {
			panic(err)
		}
	}

	fmt.Printf(
		"producer: \tjust sent %d requests with %d data objects each\n",
		REQUEST_COUNT,
		DATA_COUNT,
	)
}

func Counter(incr, stop chan struct{}) {
	for i := 0; i < REQUEST_COUNT*DATA_COUNT; i++ {
		<-incr
	}
	stop <- struct{}{}
}

func Consumer(incr chan struct{}) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf(
			"consumer: \t got %s",
			r.URL.String(),
		)
		incr <- struct{}{}
	})

	http.ListenAndServe(os.Getenv("TESTING_PORT"), nil)
}

func main() {
	var (
		incr = make(chan struct{})
		stop = make(chan struct{})
	)

	go Counter(incr, stop)
	go Consumer(incr)
	Producer()

	<-stop

	println("finished")
}
