package request

import (
	"fmt"
)

type Request struct {
	Function string                 `json:"function"`
	Args     map[string]interface{} `json:"args"`
}

func New(function string) *Request {
	var r Request
	r.Function = function
	r.Args = make(map[string]interface{})
	return &r
}

func (r Request) PutString(key string, value string) {
	r.Args[key] = value
}

func (r Request) GetString(key string) (string, error) {
	value := r.Args[key]
	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r Request) PutInteger(key string, value int64) {
	r.Args[key] = value
}

func (r Request) GetInteger(key string) (int64, error) {
	v, err := r.GetNumber(key)
	return int64(v), err
}

func (r Request) PutNumber(key string, value float64) {
	r.Args[key] = value
}

func (r Request) GetNumber(key string) (float64, error) {
	value := r.Args[key]
	v, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r Request) PutBoolean(key string, value bool) {
	r.Args[key] = value
}

func (r Request) GetBoolean(key string) (bool, error) {
	value := r.Args[key]
	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}
