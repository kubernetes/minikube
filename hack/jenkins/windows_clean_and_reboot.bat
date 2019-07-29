@echo off
call :jenkins
echo waiting to see if any jobs are coming in...
timeout 30
call :jenkins
echo doing it
taskkill /IM putty.exe
taskkill /F /IM java.exe
powershell -Command "Stop-VM minikiube"
powershell -Command "Delete-VM minikube"
rmdir /S /Q C:\Users\admin\.minikube
shutdown /r

:jenkins
tasklist | find /i /n "e2e-windows-amd64.exe">NUL
if %ERRORLEVEL% == 0 exit 1
exit /B 0
