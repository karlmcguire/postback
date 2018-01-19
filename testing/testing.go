package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	Sent     = make(map[string]map[string]time.Time, 0)
	Received = make(map[string]map[string]time.Time, 0)

	REQUEST_COUNT = flag.Int(
		"request_count",
		10,
		"number of requests to send",
	)

	DATA_COUNT = flag.Int(
		"data_count",
		1,
		"number of data objects to send with each request",
	)

	TESTING_ADDR = flag.String(
		"testing_addr",
		"http://127.0.0.1:8080",
		"the http address running this program",
	)

	TESTING_PORT = flag.String(
		"testing_port",
		":8080",
		"the http port to listen on",
	)

	INGESTION_ADDR = flag.String(
		"ingestion_addr",
		"http://127.0.0.1/ingest.php",
		"the http location of the ingestion agent",
	)
)

func Producer() {
	var (
		r   *Request
		err error
	)

	for i := 0; i < *REQUEST_COUNT; i++ {
		r = NewRequest(
			"GET",
			*TESTING_ADDR+"/?req_id={req_id}&data_id={data_id}",
		)

		Sent[fmt.Sprintf("%d", i)] = make(map[string]time.Time, 0)
		Received[fmt.Sprintf("%d", i)] = make(map[string]time.Time, 0)

		for a := 0; a < *DATA_COUNT; a++ {
			r.AddData(map[string]string{
				"req_id":  fmt.Sprintf("%d", i),
				"data_id": fmt.Sprintf("%d", a),
			})

			Sent[fmt.Sprintf("%d", i)][fmt.Sprintf("%d", a)] = time.Now()
		}

		if err = r.Send(*INGESTION_ADDR); err != nil {
			panic(err)
		}
	}

	log.Printf(
		"producer: \tjust sent %d requests with %d data objects each\n",
		*REQUEST_COUNT,
		*DATA_COUNT,
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
	for i := 0; i < *REQUEST_COUNT; i++ {
		for a := 0; a < *DATA_COUNT; a++ {
			<-incr
		}
	}

	log.Printf(
		"consumer: \tjust received %d requests taking on average %v\n",
		*REQUEST_COUNT*(*DATA_COUNT),
		GetPerformance(),
	)

	stop <- struct{}{}
}

func Consumer(incr chan struct{}) {
	log.Printf(
		"consumer: \tlistening at %s\n",
		*TESTING_ADDR,
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var (
			reqId  = r.FormValue("req_id")
			dataId = r.FormValue("data_id")
		)

		Received[reqId][dataId] = time.Now()

		incr <- struct{}{}
	})

	http.ListenAndServe(*TESTING_PORT, nil)
}

func init() {
	flag.Parse()
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

	log.Print("finished")
}
