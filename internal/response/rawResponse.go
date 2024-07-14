package response

import (
	"encoding/json"
)

type RawResponse struct {
	Code    int
	Message string
	Result  json.RawMessage
}
