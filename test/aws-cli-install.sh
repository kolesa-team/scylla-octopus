#!/bin/bash

# downloads and installs awscli on a database host
# (assumes it's running on RHEL, for use with scylladb/scylla docker image)

set -e

if command -v aws &> /dev/null
then
    echo "awscli already installed"
    exit
fi

yum -y install unzip
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install
