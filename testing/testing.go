package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	REQUEST_COUNT = 1
	DATA_COUNT    = 10
)

var Performance = make(map[string]map[string]int64, 0)

func Producer() {
	var (
		r   *Request
		err error
	)

	for i := 0; i < REQUEST_COUNT; i++ {
		Performance[fmt.Sprintf("%d", i)] = make(map[string]int64, 0)

		r = NewRequest(
			"GET",
			os.Getenv("TESTING_ADDR")+"/?time={time}&req_id={req_id}&data_id={data_id}",
		)

		for a := 0; a < DATA_COUNT; a++ {
			data := map[string]string{
				"time":    fmt.Sprintf("%d", time.Now().UnixNano()),
				"req_id":  fmt.Sprintf("%d", i),
				"data_id": fmt.Sprintf("%d", a),
			}

			r.AddData(data)
		}

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

func GetPerformance() int64 {
	var (
		count int64 = 0
		total int64 = 0
	)

	for _, req := range Performance {
		for _, diff := range req {
			total = total + diff
			count++
		}
	}

	return total / count
}

func Counter(incr, stop chan struct{}) {
	for i := 0; i < REQUEST_COUNT; i++ {
		for a := 0; a < DATA_COUNT; a++ {
			<-incr
		}
	}

	fmt.Printf(
		"consumer: \tjust received %d requests taking on average %dms\n",
		REQUEST_COUNT*DATA_COUNT,
		GetPerformance(),
	)

	stop <- struct{}{}
}

func Consumer(incr chan struct{}) {
	fmt.Printf(
		"consumer: \tlistening at %s%s\n",
		os.Getenv("TESTING_ADDR"),
		os.Getenv("TESTING_PORT"),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var (
			reqId      = r.FormValue("req_id")
			dataId     = r.FormValue("data_id")
			timeString = r.FormValue("time")
			now        = time.Now().UnixNano()
			timeNano   int64
			err        error
		)

		if timeNano, err = strconv.ParseInt(timeString, 10, 64); err != nil {
			panic(err)
		}

		Performance[reqId][dataId] = now - timeNano

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
