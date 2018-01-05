@echo off
setlocal EnableExtensions EnableDelayedExpansion

if [%1] == [] (
    goto :usage
)

if [%2] == [] (
    goto :usage
)


set TARGET_PATH=%2 %3 %4 %5 %6 %7 %8 %9

:: Remove trailing spaces
for /f "tokens=* delims= " %%a in ("!TARGET_PATH!") do set "TARGET_PATH=%%a"
for /l %%a in (1,1,100) do if "!TARGET_PATH:~-1!"==" " set "TARGET_PATH=!TARGET_PATH:~0,-1!"

:: Remove trailing ; if any
if "%PATH:~-1%"==";" (
    set PATH=!PATH:~0,-1!
)

if "%1" == "add" (
    set "PATH=!PATH!;%TARGET_PATH%"
    goto :update
)

if "%1" == "remove" (
    set "PATH=!PATH:;%TARGET_PATH%=!"
    goto :update
)

:usage
echo Script to add or remove to path environment variable
echo Usage:
echo     %0 [add^|remove] path
exit /b 1

:update

call Setx PATH "!PATH!" /m
