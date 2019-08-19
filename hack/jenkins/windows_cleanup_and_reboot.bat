:: Copyright 2019 The Kubernetes Authors All rights reserved.
::
:: Licensed under the Apache License, Version 2.0 (the "License");
:: you may not use this file except in compliance with the License.
:: You may obtain a copy of the License at
::
::     http://www.apache.org/licenses/LICENSE-2.0
::
:: Unless required by applicable law or agreed to in writing, software
:: distributed under the License is distributed on an "AS IS" BASIS,
:: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
:: See the License for the specific language governing permissions and
:: limitations under the License.

:: Periodically cleanup and reboot if no Jenkins subprocesses are running.

@echo off
call :jenkins
echo waiting to see if any jobs are coming in...
timeout 30
call :jenkins
echo doing it
taskkill /IM putty.exe
taskkill /F /IM java.exe
powershell -Command "Stop-VM minikube"
powershell -Command "Delete-VM minikube"
rmdir /S /Q C:\Users\admin\.minikube
shutdown /r

:jenkins
tasklist | find /i /n "e2e-windows-amd64.exe">NUL
if %ERRORLEVEL% == 0 exit 1
exit /B 0
