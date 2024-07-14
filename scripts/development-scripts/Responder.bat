@echo off

setlocal
cd %~dp0

responder.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
