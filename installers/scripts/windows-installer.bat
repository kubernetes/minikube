set MINIKUBE_INSTALL_HOME="$(./bin/minikube home)"
mkdir -p %MINIKUBE_INSTALL_HOME%/bin

copy ./bin/minikube-windows-amd64.exe %MINIKUBE_INSTALL_HOME%/bin/minikube.exe
copy -r ./addons %MINIKUBE_INSTALL_HOME%
copy -r ./iso %MINIKUBE_INSTALL_HOME%/cache
copy -r ./localkube %MINIKUBE_INSTALL_HOME%/cache

for %%p in (minikube.exe) do set "progpath=%%~$PATH:p"
if not defined progpath (
   rem The path to minikube.exe doesn't exist in PATH variable, insert it:
   set "PATH=%PATH%;%MINIKUBE_INSTALL_HOME%/bin/minikube.exe"
)