@echo off

setlocal
cd %~dp0

echo on
CalculatorRequest.exe -username %MQTT_USERNAME% -password %MQTT_PASSWORD% -operation div -param1 10 -param2 0
