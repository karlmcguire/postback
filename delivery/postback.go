package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-redis/redis"
)

var (
	ErrInvalidParams = errors.New("url params aren't in '{param}' format")
)

type Postback struct {
	Method string            `json:"method"`
	Url    string            `json:"url"`
	Count  int               `json:"count"`
	Params map[string]string `json:"-"`
}

func NewPostback(db *redis.Client, key string) {
	var (
		postback *Postback = &Postback{}
		value    []byte
		err      error
	)

	// get the json object from the postback:[uuid] key in redis
	if value, err = db.Get(key).Bytes(); err != nil {
		panic(err)
	}

	// delete the postback:[uuid] key from redis
	if err = db.Del(key).Err(); err != nil {
		panic(err)
	}

	// unmarshal the json object into postback variable
	if err = postback.UnmarshalJSON(value); err != nil {
		panic(err)
	}

	// parse the postback url for params
	if err = postback.Parse(nil); err != nil {
		panic(err)
	}

	// start listening for data objects on postback:[uuid]:data
	postback.Listen(db, key+":data")
}

func (p *Postback) Listen(db *redis.Client, key string) {
	var (
		data []string
		err  error
	)

	// the count field added by the ingestion agent allows this goroutine to
	// only live for as long as there are still data objects to receive
	for i := 0; i < p.Count; i++ {
		// block until data object is available
		if data, err = db.BLPop(0, key).Result(); err != nil {
			panic(err)
		}

		// handle the data object
		//go p.Respond(data[1])
		p.Respond(data[1])
	}
}

func (p *Postback) Respond(value string) {
	var (
		params map[string]string
		err    error
	)

	if err = json.Unmarshal([]byte(value), &params); err != nil {
		panic(err)
	}

	println(p.Fill(params))
}

func (p *Postback) Parse(defaults map[string]string) error {
	// check if there's an odd number of '{' or '}', and if so, return an error
	// because there's something very wrong
	if strings.Count(p.Url, "{")%2 != 0 ||
		strings.Count(p.Url, "}")%2 != 0 {
		return ErrInvalidParams
	}

	// params is a string slice containing each '{param}' in the url field
	params := regexp.MustCompile("\\{([^}]+)\\}").FindAllString(p.Url, -1)
	// the Params map is used in Fill() to iterate over each param
	p.Params = make(map[string]string, len(params))

	// copy each {param} from string slice to the Params map
	for _, param := range params {
		// if defaults contains a definition for the param, include it as the
		// value for the param, otherwise it's just empty
		p.Params[param] = defaults[param[1:len(param)-1]]
	}

	return nil
}

func (p *Postback) Fill(values map[string]string) string {
	// create a copy of p.Url to be modified
	var filled string = p.Url

	// for every param defined in p.Url
	for param, paramDefault := range p.Params {
		if value, exists := values[param[1:len(param)-1]]; exists {
			// the values map contains a definition for the param, so put it in
			// the new url
			filled = strings.Replace(filled, param, url.QueryEscape(value), -1)
		} else {
			// the values map doesn't contain a definition for the param, so put
			// the default value in the new url
			filled = strings.Replace(filled, param, url.QueryEscape(paramDefault), -1)
		}
	}

	return filled
}

func (p *Postback) UnmarshalJSON(value []byte) error {
	return json.Unmarshal(value, p)
}
