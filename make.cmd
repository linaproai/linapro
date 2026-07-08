@echo off
REM Purpose: Windows command wrapper for LinaPro's linactl tool.
REM It forwards all arguments to hack\tools\linactl.
setlocal
pushd "%~dp0hack\tools\linactl" || exit /b 1
go run . %*
set EXIT_CODE=%ERRORLEVEL%
popd
exit /b %EXIT_CODE%
