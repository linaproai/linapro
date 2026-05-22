@echo off
REM Purpose: Windows command wrapper for LinaPro's linactl tool, forwarding all arguments to hack\tools\linactl.
REM 用途：作为 LinaPro 的 Windows 命令包装入口，将所有参数转发给 hack\tools\linactl。
setlocal
pushd "%~dp0hack\tools\linactl" || exit /b 1
go run . %*
set EXIT_CODE=%ERRORLEVEL%
popd
exit /b %EXIT_CODE%
