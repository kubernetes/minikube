package cluster

var startCommand = `
sudo killall localkube || true
# Download and install localkube, if it doesn't exist yet.
if [ ! -e /usr/local/bin/localkube2 ]; then
	sudo curl -L %s -o /usr/local/bin/localkube
	sudo chmod a+x /usr/local/bin/localkube;
fi
# Fetch easy-rsa.
sudo mkdir -p /srv/kubernetes/certs && sudo chmod -R 777 /srv
if [ ! -e easy-rsa.tar.gz ]; then
	curl -L -O https://storage.googleapis.com/kubernetes-release/easy-rsa/easy-rsa.tar.gz
fi
rm -rf easy-rsa-master
tar xzf easy-rsa.tar.gz
# Create certs.
cert_ip=$(ip addr show ${interface} | grep 192.168 | sed -nEe 's/^[ \t]*inet[ \t]*([0-9.]+)\/.*$/\1/p')
ts=$(date +%%s)
if ! grep $cert_ip /srv/kubernetes/certs/kubernetes-master.crt; then
	cd easy-rsa-master/easyrsa3
	./easyrsa init-pki
	./easyrsa --batch "--req-cn=$cert_ip@$ts" build-ca nopass
	./easyrsa --subject-alt-name="IP:$cert_ip" build-server-full kubernetes-master nopass
	./easyrsa build-client-full kubecfg nopass
	cp -p pki/ca.crt /srv/kubernetes/certs/
	cp -p pki/issued/kubecfg.crt /srv/kubernetes/certs/
	cp -p pki/private/kubecfg.key /srv/kubernetes/certs/
	cp -p pki/issued/kubernetes-master.crt /srv/kubernetes/certs/
	cp -p pki/private/kubernetes-master.key /srv/kubernetes/certs/
fi
# Drop this once we get the containerized flag in.
sudo ln -s / /rootfs
# Run with nohup so it stays up. Redirect logs to useful places.
PATH=/usr/local/sbin:$PATH nohup sudo /usr/local/bin/localkube start > /var/log/localkube.out 2> /var/log/localkube.err < /dev/null &
`
