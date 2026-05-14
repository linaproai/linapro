@echo off
setlocal
pushd "%~dp0hack\tools\linactl" || exit /b 1
go run . %*
set EXIT_CODE=%ERRORLEVEL%
popd
exit /b %EXIT_CODE%
