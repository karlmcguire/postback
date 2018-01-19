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

var (
	Sent     = make(map[string]map[string]time.Time, 0)
	Received = make(map[string]map[string]time.Time, 0)
)

func Producer() {
	var (
		r   *Request
		err error
	)

	for i := 0; i < REQUEST_COUNT; i++ {
		r = NewRequest(
			"GET",
			os.Getenv("TESTING_ADDR")+"/?req_id={req_id}&data_id={data_id}",
		)

		Sent[fmt.Sprintf("%d", i)] = make(map[string]time.Time, 0)

		for a := 0; a < DATA_COUNT; a++ {
			r.AddData(map[string]string{
				"req_id":  fmt.Sprintf("%d", i),
				"data_id": fmt.Sprintf("%d", a),
			})

			Sent[fmt.Sprintf("%d", i)][fmt.Sprintf("%d", a)] = time.Now()
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

func GetPerformance() time.Duration {
	var durations []time.Duration

	for reqId, _ := range Sent {
		for dataId, _ := range Sent[reqId] {
			durations = append(
				durations,
				Received[reqId][dataId].Sub(
					Sent[reqId][dataId],
				),
			)
		}
	}

	var (
		d time.Duration
		i int
		t int64
	)

	for i, d = range durations {
		t = t + int64(d/time.Millisecond)
	}

	d, _ = time.ParseDuration(fmt.Sprintf("%dms", t/int64(i)))

	return d
}

func Counter(incr, stop chan struct{}) {
	for i := 0; i < REQUEST_COUNT; i++ {
		for a := 0; a < DATA_COUNT; a++ {
			<-incr
		}
	}

	fmt.Printf(
		"consumer: \tjust received %d requests taking on average %v\n",
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
			reqId  = r.FormValue("req_id")
			dataId = r.FormValue("data_id")
		)

		Received[reqId][dataId] = time.Now()

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
