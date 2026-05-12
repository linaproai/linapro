@echo off
go run ./hack/tools/linactl %*
exit /b %ERRORLEVEL%
