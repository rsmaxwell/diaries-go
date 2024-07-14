package main

import (
	"log/slog"
	"net/http"

	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

type GetPagesHandler struct {
}

func (h *GetPagesHandler) Handle(req request.Request) (*response.Response, bool, error) {
	slog.Debug("GetPagesHandler")

	resp := response.New(http.StatusOK)
	resp.PutString("result", "[ 'one', 'two', 'three' ]")
	return resp, false, nil
}
