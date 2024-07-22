@echo off

setlocal
cd %~dp0

echo on
CalculatorRequest.exe -operation div -param1 10 -param2 0
