@echo off

setlocal
cd %~dp0

echo on
GetPagesRequest.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
