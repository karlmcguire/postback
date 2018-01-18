package main

import (
	"encoding/json"
)

type Postback struct {
	Method string            `json:"method"`
	Url    string            `json:"url"`
	Count  int               `json:"count"`
	Params map[string]string `json:"-"`
}

func (p *Postback) UnmarshalJSON(value []byte) error {
	return json.Unmarshal(value, p)
}
