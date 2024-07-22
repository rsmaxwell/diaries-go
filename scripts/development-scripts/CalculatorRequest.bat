@echo off

setlocal
cd %~dp0

echo on
CalculatorRequest.exe -operation mul -param1 10 -param2 5
