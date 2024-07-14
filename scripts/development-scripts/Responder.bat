@echo off

setlocal
cd %~dp0

echo on
responder.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
