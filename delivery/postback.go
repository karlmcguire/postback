package main

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
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

func (p *Postback) UnmarshalJSON(value []byte) error {
	return json.Unmarshal(value, p)
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
