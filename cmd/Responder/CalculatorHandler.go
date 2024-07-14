package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

type CalculatorHandler struct {
}

func (h *CalculatorHandler) Handle(req request.Request) (resp *response.Response, quit bool, err error) {
	slog.Debug("CalculatorHandler")

	operation, err := req.GetString("operation")
	if err != nil {
		resp := response.New(http.StatusBadRequest)
		resp.PutMessage(fmt.Sprintf("could not find 'operation' in arguments: %s", err))
		return resp, false, nil
	}

	param1, err := req.GetInteger("param1")
	if err != nil {
		resp := response.New(http.StatusBadRequest)
		resp.PutMessage(fmt.Sprintf("could not find 'param1' in arguments: %s", err))
		return resp, false, nil
	}

	param2, err := req.GetInteger("param2")
	if err != nil {
		resp := response.New(http.StatusBadRequest)
		resp.PutMessage(fmt.Sprintf("could not find 'param2' in arguments: %s", err))
		return resp, false, nil
	}

	defer func() {
		if r := recover(); r != nil {
			errorText := fmt.Sprintf("%s", r)
			log.Println("RECOVER", errorText)
			resp = response.New(http.StatusBadRequest)
			resp.PutMessage(errorText)
			quit = false
			err = nil
		}
	}()

	var value int64

	switch operation {
	case "add":
		value = param1 + param2
	case "mul":
		value = param1 * param2
	case "div":
		value = param1 / param2
	case "sub":
		value = param1 - param2
	}

	resp = response.New(http.StatusOK)
	resp.PutInteger("result", value)
	return resp, false, nil
}
