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
	"log"
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
	"github.com/rsmaxwell/diaries/internal/config"
	"github.com/rsmaxwell/diaries/internal/loggerlevel"
	"github.com/rsmaxwell/diaries/internal/request"
	"github.com/rsmaxwell/diaries/internal/response"
)

const (
	qos          = 0
	requestTopic = "request"
)

func main() {

	log.Printf("QuitRequest")

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
		OnConnectError:    func(err error) { log.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			OnClientError: func(err error) { log.Printf("requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Printf("requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Printf("requested disconnect; reason code: %d\n", d.ReasonCode)
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
			log.Printf("requestor failed to subscribe (%s). This is likely to mean no messages will be received.", err)
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
		log.Fatalf("requestor failed to connect & subscribe: %s", err)
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

	r := request.New("quit")
	r.PutBoolean("quit", true)

	j, err := json.Marshal(r)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	log.Printf("Sending request: %s", j)
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
		slog.Error(fmt.Sprintf("could not decode response: %v", err))
	}

	// Handle the response
	if resp.Ok() {
		slog.Info("Responder is quitting")
	} else {
		code, _ := resp.GetCode()
		message, _ := resp.GetMessage()
		slog.Error(fmt.Sprintf("error response: code: %d, message: %s", code, message))
	}
}
