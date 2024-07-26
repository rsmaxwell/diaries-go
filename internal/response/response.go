package response

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"

	"github.com/rsmaxwell/diaries/internal/buildinfo"
)

type Response map[string]interface{}

func New(code int) *Response {
	r := make(Response)
	r["code"] = code
	return &r
}

func (r *Response) Ok() bool {
	n, err := r.GetInteger("code")
	if err != nil {
		return false
	}
	return n == http.StatusOK
}

func (r *Response) PutCode(code int) {
	(*r)["code"] = code
}

func (r *Response) GetCode() (int, error) {
	n, err := r.GetInteger("code")
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *Response) PutMessage(message string) {
	(*r)["message"] = message
}

func (r *Response) GetMessage() (string, error) {
	s, err := r.GetString("message")
	if err != nil {
		return "", err
	}
	return s, nil
}

func BadRequest(message string) *Response {
	r := make(Response)
	slog.Info(message)
	r.PutCode(http.StatusBadRequest)
	r.PutMessage(message)
	return &r
}

func InternalServerError(message string) *Response {
	r := make(Response)
	slog.Info(message)
	r.PutCode(http.StatusInternalServerError)
	r.PutMessage(message)
	return &r
}

func (r *Response) PutBuildInfo(value *buildinfo.BuildInfo) {
	(*r)["version"] = value.Version
	(*r)["buildDate"] = value.BuildDate
	(*r)["gitCommit"] = value.GitCommit
	(*r)["gitBranch"] = value.GitBranch
	(*r)["gitUrl"] = value.GitURL
}

func (r *Response) GetBuildInfo() (*buildinfo.BuildInfo, error) {

	info := new(buildinfo.BuildInfo)

	value := (*r)["version"]
	v, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %+v", value)
	}
	info.Version = v

	value = (*r)["buildDate"]
	v, ok = value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %+v", value)
	}
	info.BuildDate = v

	value = (*r)["gitCommit"]
	v, ok = value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %+v", value)
	}
	info.GitCommit = v

	value = (*r)["gitBranch"]
	v, ok = value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %+v", value)
	}
	info.GitBranch = v

	value = (*r)["gitUrl"]
	v, ok = value.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %+v", value)
	}
	info.GitURL = v

	return info, nil
}

func (r *Response) PutString(key string, value string) {
	(*r)[key] = value
}

func (r *Response) GetString(key string) (string, error) {
	value := (*r)[key]
	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r *Response) PutInteger(key string, value int64) {
	(*r)[key] = value
}

func (r *Response) GetInteger(key string) (int64, error) {
	n, err := r.GetNumber(key)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(n)), nil
}

func (r *Response) PutNumber(key string, value float64) {
	(*r)[key] = value
}

func (r *Response) GetNumber(key string) (float64, error) {
	value := (*r)[key]
	v, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r *Response) PutBoolean(key string, value bool) {
	(*r)[key] = value
}

func (r *Response) GetBoolean(key string) (bool, error) {
	value := (*r)[key]
	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}
