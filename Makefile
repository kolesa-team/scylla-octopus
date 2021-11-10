.PHONY: test

APP_NAME := scylla-octopus
PACKAGE := github.com/kolesa-team/$(APP_NAME)
GIT_REV ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION ?= v1.0.0

BUILD_LDFLAGS := "-X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT_REV} -X ${PACKAGE}/cmd.buildDate=${DATE}"

DOCKER_IMAGE_PATH := kolesa-team/scylla-octopus
DOCKER_IMAGE_TAG ?= ${VERSION}
DOCKER_IMAGE := ${DOCKER_IMAGE_PATH}:${DOCKER_IMAGE_TAG}

build:
	CGO_ENABLED=0 go build -mod=mod \
		-o dist/${APP_NAME} \
		-ldflags ${BUILD_LDFLAGS}

test:
	go test ./... -v

# builds docker image
docker-image:
	docker build -t ${DOCKER_IMAGE} .

# prepares a database node from docker-compose for development.
# example: make prepare-test-node node=scylla-node1
prepare-test-node: add-ssh-key install-awscli install-pigz

# creates a database with example dataset for testing.
# example: make init-db node=scylla-node1
init-db:
	docker-compose exec $(node) cqlsh -f /test/database.cql

# installs awscli on a given node
# example: make install-awscli node=scylla-node1
install-awscli:
	docker-compose exec $(node) /test/aws-cli-install.sh

# install pigz for compress backup
# example: make install-pigz node=scylla-node1
install-pigz:
	docker-compose exec $(node) /test/pigz-install.sh

# adds an SSH key to a given node,
# so that scylla-octopus can log in over SSH
add-ssh-key:
	docker-compose exec $(node) sh -c 'mkdir -p -m700 /root/.ssh && cat /test/ssh/id_rsa.pub >> /root/.ssh/authorized_keys'

# destroys a docker-compose testing environment.
# TODO this should also remove test/node*/data, test/node*/commitlog.
teardown:
	docker-compose stop
	docker-compose down --rmi local --volumes --remove-orphans
