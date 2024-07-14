@echo off

setlocal
cd %~dp0

echo on
QuitRequest.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD%
