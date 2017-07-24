set -e
MINIKUBE_INSTALL_HOME="$(./bin/minikube home)"
mkdir -p ${MINIKUBE_INSTALL_HOME}/bin
cp ./bin/docker-machine-driver-kvm ${MINIKUBE_INSTALL_HOME}/bin
cp ./bin/minikube-darwin-amd64 ${MINIKUBE_INSTALL_HOME}/bin/minikube

cp -r ./addons ${MINIKUBE_INSTALL_HOME}
cp -r ./iso ${MINIKUBE_INSTALL_HOME}/cache
cp -r ./localkube ${MINIKUBE_INSTALL_HOME}/cache

MINIKUBE_PATH_COMMENT='# The next line updates PATH for minikube.'
PATH_CONFIG=${HOME}/.bash_profile
touch ${PATH_CONFIG}
if grep -q "${MINIKUBE_PATH_COMMENT}" ${PATH_CONFIG}
  then
    echo '' >>${PATH_CONFIG}
    echo ${MINIKUBE_PATH_COMMENT} >>${PATH_CONFIG}
    echo 'if [ -f "'${MINIKUBE_INSTALL_HOME}'/path.bash.inc" ]; then source "'${MINIKUBE_INSTALL_HOME}'/path.bash.inc"; fi' >>${PATH_CONFIG}
    echo 'Reload your current session for the change made to the libvirtd to take effect.  This is required to use minikube with kvm'
fi
