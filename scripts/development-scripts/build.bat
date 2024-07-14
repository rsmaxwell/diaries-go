@echo off

setlocal
cd %~dp0\..\..

echo on
go install ./...
