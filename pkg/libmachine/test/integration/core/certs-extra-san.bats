#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_disposable_machine

@test "$DRIVER: create" {
  run machine create --tls-san foo.bar.tld --tls-san 10.42.42.42  -d $DRIVER $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: verify that server cert contains the extra SANs" {
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'DNS:foo.bar.tld'
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'IP Address:10.42.42.42'
}

@test "$DRIVER: verify that server cert SANs are still there after 'regenerate-certs'" {
    machine regenerate-certs -f $NAME
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'DNS:foo.bar.tld'
    machine ssh $NAME -- openssl x509 -in /var/lib/boot2docker/server.pem -text | grep 'IP Address:10.42.42.42'
}
