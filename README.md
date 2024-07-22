# diaries-go

# Overview

Some webapps need both a request/resonse and a publish/subscribe messaging model, to both make requests to a server which require a particular response (e.g. logon) and also to subscribe to events as thy happen (e.g. display share price). http works well for request/resonse but has to resort to polling to keep uptodate in a dynamic environment. MQTT was built around the publish/subscribe model and no in v5 supports request/resonse as well. This package uses these mqtt v5 features to provide an implementation of request/resonse messaging. 

*MQTT* requires a broker process to be running to which clients connect. In the case of mqtt-rpc-go there is one client (*Responder*) which listens to a well known topic, and sends replies to a reply_topic, the name of which was contained in the original request. There may be many *Requester* clients which send requests, and wait for a reply.  *Requester* clients can only make requests which are supported by the *Responder*.

The *MQTT* broker used in this app is **Mosquitto** for the broker. Connections to the *MQTT* broker, which will be exposed to the internet, need to be appropiatly secured

# Windows Setup

## Mosquitto - MQTT Broker
 
 - Install and run Mosquitto from https://mosquitto.org

## Build

 - Install GO from https://go.dev/doc/install

 - Set GOPATH to your installation, e.g. set GOPATH=%USERPROFILE%/go

 - Make sure that your PATH includes: %GOPATH%\bin

 - Run %PROJECT_DIR%\scripts\build, (which will build and install the server and client to:  %GOPATH%\bin )


## Responder

 - Run %PROJECT_DIR%\scripts\development-scripts\diary-server

## Requests

 - Start a new terminal (i.e. cmd.exe ) in the %PROJECT_DIR%\scripts\development-scripts directory

 - Run CalculatorRequest.bat
 - Run GetpagesRequest.bat
 - Run QuitRequest.bat


## Notes

This issue adds a timeout on the rpc request

https://github.com/eclipse/paho.golang/issues/108
https://github.com/eclipse/paho.golang/pull/260
