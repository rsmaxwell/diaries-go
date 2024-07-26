package main

import (
	"log/slog"
	"net/http"

	"github.com/rsmaxwell/diaries/internal/request"
	"github.com/rsmaxwell/diaries/internal/response"
)

type GetPagesHandler struct {
}

func (h *GetPagesHandler) Handle(req request.Request) (*response.Response, bool, error) {
	slog.Debug("GetPagesHandler")

	resp := response.New(http.StatusOK)
	resp.PutString("result", "[ 'one', 'two', 'three' ]")
	return resp, false, nil
}
