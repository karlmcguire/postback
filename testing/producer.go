package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Request is the type sent to the Ingestion Agent in JSON.
type Request struct {
	Endpoint Endpoint            `json:"endpoint"`
	Data     []map[string]string `json:"data"`
}

// Endpoint is a JSON object send with Request.
type Endpoint struct {
	Method string `json:"method"`
	Url    string `json:"url"`
}

func NewRequest(method, url string) *Request {
	return &Request{
		Endpoint: Endpoint{
			Method: method,
			Url:    url,
		},
		Data: make([]map[string]string, 0),
	}
}

func (r *Request) AddData(data map[string]string) {
	r.Data = append(r.Data, data)
}

func (r *Request) Send(url string) error {
	var (
		b   []byte
		err error
	)

	if b, err = json.Marshal(r); err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)

	if _, err = http.Post(url, "application/json", buf); err != nil {
		return err
	}

	return nil
}
