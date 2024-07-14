package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

type QuitHandler struct {
}

func (h *QuitHandler) Handle(req request.Request) (*response.Response, bool, error) {
	slog.Debug("QuitHandler")

	quit, err := req.GetBoolean("quit")
	if err != nil {
		resp := response.New(http.StatusBadRequest)
		resp.PutMessage(fmt.Sprintf("could not find 'quit' in arguments: %s", err))
		return resp, false, nil
	}

	resp := response.New(http.StatusOK)
	return resp, quit, nil
}
