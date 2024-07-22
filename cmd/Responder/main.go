/* see:
 *    https://github.com/eclipse/paho.golang/blob/v0.21.0/autopaho/examples/basics/basics.go
 *    https://github.com/eclipse/paho.golang/blob/master/autopaho/examples/rpc/main.go
 */

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/config"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/loggerlevel"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"

	_ "github.com/lib/pq"
)

const (
	qos          = 0
	requestTopic = "request"
)

type Handler interface {
	Handle(request.Request) (*response.Response, bool, error)
}

var (
	requestHandlers = map[string]Handler{
		"buildinfo":  new(BuildInfoHandler),
		"calculator": new(CalculatorHandler),
		"getPages":   new(GetPagesHandler),
		"quit":       new(QuitHandler),
	}
)

func ConnectDatabase(dBConfig *config.DBConfig) (*sql.DB, error) {

	driverName := dBConfig.DriverName()
	connectionString := dBConfig.ConnectionString()

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		slog.Error(err.Error())
		slog.Error(fmt.Sprintf("driverName: %s", driverName))
		slog.Error(fmt.Sprintf("connectionString: %s", connectionString))
		return nil, fmt.Errorf("could not connect to database")
	}

	return db, err
}

func main() {

	err := loggerlevel.SetLoggerLevel()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("diaries Responder")

	config, err := config.Read()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	db, err := ConnectDatabase(&config.Db)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	serverUrl, err := url.Parse(config.Mqtt.GetServer())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mqttConfig := autopaho.ClientConfig{
		ServerUrls:        []*url.URL{serverUrl},
		KeepAlive:         30,
		ConnectRetryDelay: 2 * time.Second,
		ConnectTimeout:    5 * time.Second,
		OnConnectError:    func(err error) { slog.Info(fmt.Sprintf("error whilst attempting connection: %s\n", err)) },
		ClientConfig: paho.ClientConfig{
			OnClientError: func(err error) { slog.Info(fmt.Sprintf("requested disconnect: %s\n", err)) },

			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					slog.Info(fmt.Sprintf("requested disconnect: %s\n", d.Properties.ReasonString))
				} else {
					slog.Info(fmt.Sprintf("requested disconnect; reason code: %d\n", d.ReasonCode))
				}
			},
		},
		ConnectUsername: config.Mqtt.Username,
		ConnectPassword: []byte(config.Mqtt.Password),
	}

	mqttConfig.ClientConfig.ClientID = "listener"
	// Subscribing in OnConnectionUp is the recommended approach because this ensures the subscription is reestablished
	// following reconnection (the subscription should survive `cliCfg.SessionExpiryInterval` after disconnection,
	// but in this case that is 0, and it's safer if we don't assume the session survived anyway).
	mqttConfig.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: requestTopic, QoS: qos},
			},
		}); err != nil {
			slog.Info(fmt.Sprintf("listener failed to subscribe (%s). This is likely to mean no messages will be received.", err))
			return
		}
	}
	mqttConfig.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(received paho.PublishReceived) (bool, error) {

			slog.Info(fmt.Sprintf("Received request: %s", string(received.Packet.Payload)))

			if received.Packet.Properties == nil {
				slog.Info("discarding request with no properties")
				return true, nil
			}

			if received.Packet.Properties.CorrelationData == nil {
				slog.Info("discarding request with no CorrelationData")
				return true, nil
			}

			if received.Packet.Properties.ResponseTopic == "" {
				slog.Info("discarding request with empty responseTopic")
				return true, nil
			}

			resp, quit, err := getResult(received)
			if err != nil {
				slog.Info(err.Error())
				return true, nil
			}

			body, err := json.Marshal(resp)
			if err != nil {
				slog.Info(err.Error())
				return true, nil
			}

			slog.Info(fmt.Sprintf("Sending reply: %s", string(body)))

			_, err = received.Client.Publish(ctx, &paho.Publish{
				Properties: &paho.PublishProperties{
					CorrelationData: received.Packet.Properties.CorrelationData,
				},
				Topic:   received.Packet.Properties.ResponseTopic,
				Payload: body,
			})
			if err != nil {
				slog.Info(err.Error())
				return true, nil
			}

			if quit {
				wg.Done()
			}
			return true, nil
		}}

	_, err = autopaho.NewConnection(ctx, mqttConfig)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Wait till asked to quit
	wg.Wait()
	slog.Info("Quitting")
}

func getResult(received paho.PublishReceived) (*response.Response, bool, error) {

	var resp *response.Response
	var req request.Request
	if err := json.NewDecoder(bytes.NewReader(received.Packet.Payload)).Decode(&req); err != nil {
		resp = response.BadRequest(fmt.Sprintf("request could not be decoded: %v", err))
		return resp, false, nil
	}

	if req.Args == nil {
		resp = response.BadRequest("missing request")
		return resp, false, nil
	}

	if len(req.Function) == 0 {
		resp = response.BadRequest("empty function")
		return resp, false, nil
	}

	handler := requestHandlers[req.Function]
	if handler == nil {
		resp = response.BadRequest(fmt.Sprintf("unexpected function: %s", req.Function))
		return resp, false, nil
	}

	resp, quit, err := handler.Handle(req)
	if err != nil {
		resp = response.BadRequest(fmt.Sprintf("handler '%s' failed: %s", req.Function, err))
		return resp, false, nil
	}

	if resp == nil {
		resp = response.BadRequest("response is null")
		return resp, false, nil
	}

	return resp, quit, err
}
