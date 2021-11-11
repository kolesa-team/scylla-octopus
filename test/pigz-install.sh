#!/bin/bash

# downloads and installs pigz on a database host

set -e

if command -v pigz &> /dev/null
then
    echo "pigz already installed"
    exit
fi

yum -y install pigz
