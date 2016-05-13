#!/bin/bash

set -e

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH

KEY_PAIR_NAME=rkt-testing-${USER}
SECURITY_GROUP=rkt-testing-${USER}-security-group

## First time only
if [ "$1" = "setup" ] ; then
  MYIP=$(curl --silent http://checkip.amazonaws.com/)

  aws ec2 create-key-pair --key-name $KEY_PAIR_NAME --query 'KeyMaterial' --output text > ${KEY_PAIR_NAME}.pem
  chmod 0600 ${KEY_PAIR_NAME}.pem
  aws ec2 create-security-group --group-name $SECURITY_GROUP --description "Security group for rkt testing"
  aws ec2 authorize-security-group-ingress --group-name $SECURITY_GROUP --protocol tcp --port 22 --cidr $MYIP/32
  exit 0
fi

DISTRO=$1
GIT_URL=${2-https://github.com/coreos/rkt.git}
GIT_BRANCH=${3-master}

test -f cloudinit/${DISTRO}.cloudinit
CLOUDINIT_IN=$PWD/cloudinit/${DISTRO}.cloudinit

if [ "$DISTRO" = "fedora-22" ] ; then
  # https://getfedora.org/en/cloud/download/
  # Search on AWS or look at
  # https://apps.fedoraproject.org/datagrepper/raw?category=fedimg
  # Sources: https://github.com/fedora-infra/fedimg/blob/develop/bin/list-the-amis.py

  # Fedora-Cloud-Base-22-20151026.x86_64-eu-central-1-HVM-standard-0 was deleted

  # Fedora-Cloud-Base-22-20150521.x86_64-eu-central-1-HVM-standard-0
  AMI=ami-a88eb0b5
  AWS_USER=fedora
elif [ "$DISTRO" = "fedora-23" ] ; then
  # Fedora-Cloud-Base-23-20160129.x86_64-eu-central-1-HVM-standard-0
  AMI=ami-4d3e2621
  AWS_USER=fedora
elif [ "$DISTRO" = "fedora-rawhide" ] ; then
  # rawhide is currently broken
  # Error: nothing provides libpsl.so.0()(64bit) needed by wget-1.17.1-1.fc24.x86_64

  # Fedora-Cloud-Base-rawhide-20160127.x86_64-eu-central-1-HVM-standard-0
  AMI=ami-877068eb
  AWS_USER=fedora
elif [ "$DISTRO" = "ubuntu-1604" ] ; then
  # https://cloud-images.ubuntu.com/locator/ec2/
  # ubuntu/images-milestone/hvm/ubuntu-xenial-alpha2-amd64-server-20160125
  AMI=ami-b4a5b9d8
  AWS_USER=ubuntu
elif [ "$DISTRO" = "ubuntu-1510" ] ; then
  # https://cloud-images.ubuntu.com/locator/ec2/
  # ubuntu/images/hvm/ubuntu-wily-15.10-amd64-server-20160123
  AMI=ami-e9869f85
  AWS_USER=ubuntu
elif [ "$DISTRO" = "debian" ] ; then
  # https://wiki.debian.org/Cloud/AmazonEC2Image/Jessie
  # Debian 8.1
  AMI=ami-02b78e1f
  AWS_USER=admin
elif [ "$DISTRO" = "centos" ] ; then
  # Needs to subscribe first, see:
  # https://wiki.centos.org/Cloud/AWS
  # CentOS-7 x86_64 HVM
  AMI=ami-e68f82fb
  AWS_USER=centos
fi

test -n "$AMI"
test -n "$AWS_USER"
test -f "${KEY_PAIR_NAME}.pem"

CLOUDINIT=$(mktemp --tmpdir rkt-cloudinit.XXXXXXXXXX)
sed -e "s#@GIT_URL@#${GIT_URL}#g" \
    -e "s#@GIT_BRANCH@#${GIT_BRANCH}#g" \
    < $CLOUDINIT_IN >> $CLOUDINIT

INSTANCE_ID=$(aws ec2 run-instances \
	--image-id $AMI \
	--count 1 \
	--key-name $KEY_PAIR_NAME \
	--security-groups $SECURITY_GROUP \
	--instance-type t2.micro \
	--instance-initiated-shutdown-behavior terminate \
	--user-data file://$CLOUDINIT \
	--output text \
	--query 'Instances[*].InstanceId' \
	)
echo INSTANCE_ID=$INSTANCE_ID

while state=$(aws ec2 describe-instances \
	--instance-ids $INSTANCE_ID \
	--output text \
	--query 'Reservations[*].Instances[*].State.Name' \
	); test "$state" = "pending"; do
  sleep 1; echo -n '.'
done; echo " $state"

AWS_IP=$(aws ec2 describe-instances \
	--instance-ids $INSTANCE_ID \
	--output text \
	--query 'Reservations[*].Instances[*].PublicIpAddress' \
	)
echo AWS_IP=$AWS_IP

rm -f $CLOUDINIT

sleep 5
aws ec2 get-console-output --instance-id $INSTANCE_ID --output text |
  perl -ne 'print if /BEGIN SSH .* FINGERPRINTS/../END SSH .* FINGERPRINTS/'

echo "Connect with:"
echo ssh -o ServerAliveInterval=20 -i ${SCRIPTPATH}/${KEY_PAIR_NAME}.pem ${AWS_USER}@${AWS_IP}
echo "Check the logs with:"
echo tail -n 5000 -f /var/tmp/rkt-test.log

