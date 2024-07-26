package main

import (
	"log/slog"
	"net/http"

	"github.com/rsmaxwell/diaries/internal/buildinfo"
	"github.com/rsmaxwell/diaries/internal/request"
	"github.com/rsmaxwell/diaries/internal/response"
)

type BuildInfoHandler struct {
}

func (h *BuildInfoHandler) Handle(req request.Request) (*response.Response, bool, error) {
	slog.Debug("BuildInfoHandler")

	info := buildinfo.NewBuildInfo()

	r := response.New(http.StatusOK)
	r.PutBuildInfo(info)
	return r, false, nil
}
