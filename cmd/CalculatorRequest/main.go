/* see:
 *    https://github.com/eclipse/paho.golang/blob/v0.21.0/autopaho/examples/basics/basics.go
 *    https://github.com/eclipse/paho.golang/blob/master/autopaho/examples/rpc/main.go
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/autopaho/extensions/rpc"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/loggerlevel"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

const qos = 0

func main() {

	slog.Info("CalculatorRequest")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := flag.String("server", "mqtt://127.0.0.1:1883", "The URL of the MQTT server")
	rTopic := flag.String("rtopic", "request", "Topic for requests to go to")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	operation := flag.String("operation", "", "The calculation operation (add, sub, mul, div)")
	param1Flag := flag.String("param1", "", "The first integer argument")
	param2Flag := flag.String("param2", "", "The second integer argument")
	flag.Parse()

	err := loggerlevel.SetLoggerLevel()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	serverUrl, err := url.Parse(*server)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	param1, err := strconv.ParseInt(*param1Flag, 10, 64)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	param2, err := strconv.ParseInt(*param2Flag, 10, 64)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	config := autopaho.ClientConfig{
		ServerUrls:        []*url.URL{serverUrl},
		KeepAlive:         30,
		ConnectRetryDelay: 2 * time.Second,
		ConnectTimeout:    5 * time.Second,
		OnConnectError:    func(err error) { slog.Error(fmt.Sprintf("error whilst attempting connection: %s\n", err)) },
		ClientConfig: paho.ClientConfig{
			OnClientError: func(err error) { slog.Error(fmt.Sprintf("requested disconnect: %s\n", err)) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					slog.Error(fmt.Sprintf("requested disconnect: %s\n", d.Properties.ReasonString))
				} else {
					slog.Error(fmt.Sprintf("requested disconnect; reason code: %d\n", d.ReasonCode))
				}
			},
		},
		ConnectUsername: *username,
		ConnectPassword: []byte(*password),
	}

	config.ClientConfig.ClientID = "requester"

	initialSubscriptionMade := make(chan struct{}) // Closed when subscription made (otherwise we might send request before subscription in place)
	var initialSubscriptionOnce sync.Once          // We only want to close the above once!

	config.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()

		// Subscribe to the responseTopic
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: fmt.Sprintf("response/%s", config.ClientID), QoS: qos},
			},
		}); err != nil {
			slog.Warn(fmt.Sprintf("requestor failed to subscribe (%s). This is likely to mean no messages will be received.", err))
			return
		}
		initialSubscriptionOnce.Do(func() { close(initialSubscriptionMade) })
	}

	router := paho.NewStandardRouter()
	config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(p paho.PublishReceived) (bool, error) {
			router.Route(p.Packet.Packet())
			return false, nil
		}}

	cm, err := autopaho.NewConnection(ctx, config)
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
		return
	case <-initialSubscriptionMade:
	}

	h, err := rpc.NewHandler(ctx, rpc.HandlerOpts{
		Conn:             cm,
		Router:           router,
		ResponseTopicFmt: "response/%s",
		ClientID:         config.ClientID,
	})

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	r := request.New("calculator")
	r.PutString("operation", *operation)
	r.PutInteger("param1", param1)
	r.PutInteger("param2", param2)

	j, err := json.Marshal(r)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Sending request: %s", j))
	reply, err := h.Request(ctx, &paho.Publish{
		Topic:   *rTopic,
		Payload: []byte(j),
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// log.Printf("Received response: %s", string(resp.Payload))

	var resp response.Response
	if err := json.NewDecoder(bytes.NewReader(reply.Payload)).Decode(&resp); err != nil {
		log.Printf("could not decode response: %v", err)
	}

	// Handle the response
	if resp.Ok() {
		result, _ := resp.GetInteger("result")
		log.Printf("result: %d", result)
	} else {
		code, _ := resp.GetCode()
		message, _ := resp.GetMessage()
		log.Printf("ERROR: code: %d, message: %s", code, message)
	}
}
