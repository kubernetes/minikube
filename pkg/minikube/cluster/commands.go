package cluster

import "fmt"

var startCommand = `sudo killall localkube || true
# Download and install localkube, if it doesn't exist yet.
if [ ! -e /usr/local/bin/localkube ]; then
	sudo curl --compressed -L %s -o /usr/local/bin/localkube
	sudo chmod a+x /usr/local/bin/localkube
fi
# Run with nohup so it stays up. Redirect logs to useful places.
PATH=/usr/local/sbin:$PATH nohup sudo /usr/local/bin/localkube start > /var/log/localkube.out 2> /var/log/localkube.err < /dev/null &
`

func getStartCommand(localkubeURL string) string {
	return fmt.Sprintf(startCommand, localkubeURL)
}
