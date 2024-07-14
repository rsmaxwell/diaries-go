@echo off

setlocal
cd %~dp0

echo on
BuildInfoRequest.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
