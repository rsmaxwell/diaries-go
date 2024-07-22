/* see:
 *    https://github.com/eclipse/paho.golang/blob/v0.21.0/autopaho/examples/basics/basics.go
 *    https://github.com/eclipse/paho.golang/blob/master/autopaho/examples/rpc/main.go
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/autopaho/extensions/rpc"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/config"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/loggerlevel"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

const (
	qos          = 0
	requestTopic = "request"
)

func main() {

	slog.Info("BuildInfoRequest")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config, err := config.Read()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = loggerlevel.SetLoggerLevel()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	serverUrl, err := url.Parse(config.Mqtt.GetServer())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	mqttConfig := autopaho.ClientConfig{
		ServerUrls:        []*url.URL{serverUrl},
		KeepAlive:         30,
		ConnectRetryDelay: 2 * time.Second,
		ConnectTimeout:    5 * time.Second,
		OnConnectError:    func(err error) { slog.Info(fmt.Sprintf("error whilst attempting connection: %s", err)) },
		ClientConfig: paho.ClientConfig{
			OnClientError: func(err error) { slog.Info(fmt.Sprintf("requested disconnect: %s", err)) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					slog.Info(fmt.Sprintf("requested disconnect: %s", d.Properties.ReasonString))
				} else {
					slog.Info(fmt.Sprintf("requested disconnect; reason code: %d", d.ReasonCode))
				}
			},
		},
		ConnectUsername: config.Mqtt.Username,
		ConnectPassword: []byte(config.Mqtt.Password),
	}

	mqttConfig.ClientConfig.ClientID = "requester"

	initialSubscriptionMade := make(chan struct{}) // Closed when subscription made (otherwise we might send request before subscription in place)
	var initialSubscriptionOnce sync.Once          // We only want to close the above once!

	mqttConfig.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()

		// Subscribe to the responseTopic
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: fmt.Sprintf("response/%s", mqttConfig.ClientID), QoS: qos},
			},
		}); err != nil {
			slog.Info(fmt.Sprintf("requestor failed to subscribe (%s). This is likely to mean no messages will be received.", err))
			return
		}
		initialSubscriptionOnce.Do(func() { close(initialSubscriptionMade) })
	}

	router := paho.NewStandardRouter()
	mqttConfig.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(p paho.PublishReceived) (bool, error) {
			router.Route(p.Packet.Packet())
			return false, nil
		}}

	cm, err := autopaho.NewConnection(ctx, mqttConfig)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Wait for the subscription to be made (otherwise we may miss the response!)
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case <-connCtx.Done():
		slog.Error(fmt.Sprintf("requestor failed to connect & subscribe: %s", err))
	case <-initialSubscriptionMade:
	}

	h, err := rpc.NewHandler(ctx, rpc.HandlerOpts{
		Conn:             cm,
		Router:           router,
		ResponseTopicFmt: "response/%s",
		ClientID:         mqttConfig.ClientID,
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	r := request.New("buildinfo")

	j, err := json.Marshal(r)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Sending request: %s", j))
	reply, err := h.Request(ctx, &paho.Publish{
		Topic:   requestTopic,
		Payload: []byte(j),
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var resp response.Response
	if err := json.NewDecoder(bytes.NewReader(reply.Payload)).Decode(&resp); err != nil {
		slog.Info(fmt.Sprintf("could not decode response: %v", err))
	}

	// Handle the response
	if resp.Ok() {
		info, err := resp.GetBuildInfo()
		if err != nil {
			slog.Info(fmt.Sprintf("error: %s", err.Error()))
		} else {
			slog.Info(fmt.Sprintf("Version:   %s", info.Version))
			slog.Info(fmt.Sprintf("BuildDate: %s", info.BuildDate))
			slog.Info(fmt.Sprintf("GitBranch: %s", info.GitBranch))
			slog.Info(fmt.Sprintf("GitCommit: %s", info.GitCommit))
			slog.Info(fmt.Sprintf("GitURL:    %s", info.GitURL))
		}
	} else {
		code, _ := resp.GetCode()
		message, _ := resp.GetMessage()
		slog.Info(fmt.Sprintf("error response: code: %d, message: %s", code, message))
	}
}
