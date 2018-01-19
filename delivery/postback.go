package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var (
	// ErrInvalidParams is returned by Postback.Parse() when the number of '{'
	// brackets doesn't match the number of '}' brackets.
	ErrInvalidParams = errors.New("url params aren't in '{param}' format")

	// ErrInvalidMethod is thrown when a new postback object is pulled from
	// Redis and the "method" field isn't GET, POST, or PUT.
	ErrInvalidMethod = errors.New("method field isn't a valid HTTP method")
)

// Postback is the "task" data type used so the Delivery Agent can handle
// requests forwarded to Redis by the Ingestion Agent. A Postback has one
// "endpoint" and at least one "data" object.
type Postback struct {
	// Method specifies the HTTP method to be used when making a request to Url.
	Method string `json:"method"`

	// Url contains the destination URL with {params} to be filled by data
	// objects. Here's an example:
	//
	// "http://sample.com/data?title={mascot}"
	//
	// {mascot} will be filled by a "mascot" field in the data objects for this
	// endpoint.
	Url string `json:"url"`

	// Count is the number of data objects for this endpoint.
	Count int `json:"count"`

	// Params is created when Postback.Parse() is called. It's a helper field
	// for keeping track of each {param} in the Url field.
	Params map[string]string `json:"-"`
}

// NewPostback is called when a "postback:[uuid]" value is pushed to the
// "postbacks" list on the Redis instance. This function creates a new Postback
// struct and populates each field, then calls Postback.Listen() to listen for
// data objects.
func NewPostback(db *redis.Client, key string) {
	var (
		p *Postback = &Postback{}
		// this will be the "postback:[uuid]" value
		value []byte
		err   error
	)

	// get the json object from the postback:[uuid] key in redis
	if value, err = db.Get(key).Bytes(); err != nil {
		log.Print(err)
		return
	}

	// delete the postback:[uuid] key from redis
	if err = db.Del(key).Err(); err != nil {
		log.Print(err)
		return
	}

	// unmarshal json object to p
	if err = json.Unmarshal(value, p); err != nil {
		log.Print(err)
		return
	}

	// make sure method field is valid, can add more if needed
	if p.Method != "GET" && p.Method != "POST" && p.Method != "PUT" {
		log.Print(ErrInvalidMethod)
		return
	}

	// parse the postback url for params
	if err = p.Parse(nil); err != nil {
		log.Print(err)
		return
	}

	// start listening for data objects on postback:[uuid]:data
	p.Listen(db, key+":data")
}

// Listen is called after NewPostback and accepts up to p.Count data objects
// from Redis. Each time a data object is pushed to "postback:[uuid]:data", this
// function will start a new goroutine calling p.Respond() with the data fields.
func (p *Postback) Listen(db *redis.Client, key string) {
	var (
		// will contain the data object json
		data []string
		err  error
	)

	// the count field added by the ingestion agent allows this goroutine to
	// only live for as long as there are still data objects to receive
	for i := 0; i < p.Count; i++ {
		// block until data object is available
		if data, err = db.BLPop(0, key).Result(); err != nil {
			log.Print(err)
			return
		}

		// handle the data object
		go p.Respond(data[1])
	}
}

// Respond is called for each data object. It fills p.Url with the specified
// fields and makes an HTTP request to "deliver" the data.
func (p *Postback) Respond(value string) {
	var (
		params map[string]string
		client *http.Client = &http.Client{}
		res    *http.Response
		body   []byte
		req    *http.Request
		err    error
	)

	// convert the json string into string map
	if err = json.Unmarshal([]byte(value), &params); err != nil {
		log.Print(err)
		return
	}

	// create a new request using the endpoint method/url and data params
	if req, err = http.NewRequest(p.Method, p.Fill(params), nil); err != nil {
		log.Print(err)
		return
	}

	// execute the http request
	if res, err = client.Do(req); err != nil {
		log.Print(err)
		return
	}

	// read the body
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.Print(err)
		return
	}

	log.Printf(
		"received: %s\n\tdelivered: %v\n\tstatus: %s\n\tbody: '%s'\n",
		req.URL.String(),
		time.Now(),
		res.Status,
		string(body),
	)
}

// Parse reads each {param} from p.Url into memory, so that it can easily be
// filled later.
//
// The defaults parameter contains a map where each key is a p.Url {param}
// (without brackets), and each value is the default value. For example:
//
// If p.Url is "http://sample.com/data?title={mascot}" and defaults is:
// {"mascot": "default_mascot"}, then Postback.Fill(nil) will return:
// "http://sample.com/data?title=default_mascot" if values aren't provided.
func (p *Postback) Parse(defaults map[string]string) error {
	// make sure the {params} are valid
	if strings.Count(p.Url, "{") != strings.Count(p.Url, "}") {
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

// Fill takes the values map and replaces each {param} in p.Url with it's
// associated value. If a key doesn't exist in values for each {param}, then
// the defaults will be used.
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
