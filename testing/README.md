# Testing

```
Usage of ./testing:
  -data_count int
        number of data objects to send with each request (default 1)
  -ingestion_addr string
        the http location of the ingestion agent (default "http://127.0.0.1/ingest.php")
  -request_count int
        number of requests to send (default 10)
  -testing_addr string
        the http address running this program (default "http://127.0.0.1:8080")
  -testing_port string
        the http port to listen on (default ":8080")
```

## Measured

This tool measures the duration from the initial HTTP request to the Ingestion Agent to the final HTTP request sent by the Delivery Agent.

## Performance

I ran these on the droplet provided, I'm not sure about the hardware specs. Of course, this is such a small sample that I'd take these numbers with a grain of salt. But I find them useful for getting a ballpark idea.

| Requests | Data Objects per Request | Average Time (20 runs) |
|:--------:|:------------------------:|:----------------------:|
| 1 | 10 | 14ms |
| 10 | 1 | 8ms |
| 10 | 10 | 65ms |
| 100 | 2 | 85ms |