package main

import (
	"log/slog"
	"net/http"

	"github.com/rsmaxwell/mqtt-rpc-go/internal/buildinfo"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

type BuildInfoHandler struct {
}

func (h *BuildInfoHandler) Handle(req request.Request) (*response.Response, bool, error) {
	slog.Info("BuildInfoHandler")

	info := buildinfo.NewBuildInfo()

	r := response.New(http.StatusOK)
	r.PutBuildInfo(info)
	return r, false, nil
}
