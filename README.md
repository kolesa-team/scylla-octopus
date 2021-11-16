## scylla-octopus: backup and maintenance utility for scylladb

Scylla-octopus attempts to reproduce some functionality of [Scylla Manager](https://docs.scylladb.com/operating-scylla/manager/) (which is not free) and [Medusa for Apache Cassandra](https://github.com/thelastpickle/cassandra-medusa) (which is not compatible with Scylla).

[![Actions Status](https://github.com/kolesa-team/scylla-octopus/workflows/test/badge.svg)](https://github.com/kolesa-team/scylla-octopus/actions)
[![codecov](https://codecov.io/gh/kolesa-team/scylla-octopus/branch/master/graph/badge.svg)](https://codecov.io/gh/kolesa-team/scylla-octopus)
[![Go Report Card](https://goreportcard.com/badge/github.com/kolesa-team/scylla-octopus)](https://goreportcard.com/report/github.com/kolesa-team/scylla-octopus)

### Features:

* Back up a single node or a database cluster
  * database schema export
  * snapshots of all or selected keyspaces
  * optional backup compression with `pigz`
* Upload a backup to s3-compatible storage with `awscli`
  * Backups in remote storage can be expired and removed automatically
* Database maintenance with `nodetool repair`
* Webhook support for notifications about backup completion and/or errors

Future plans:

* Backup restoration
* Configure github actions to build a docker image and run tests

----------------------

### О проекте

Scylla-octopus - утилита для бэкапа и обслуживания scylladb.
В ней реализована часть функциональности платной [Scylla Manager](https://docs.scylladb.com/operating-scylla/manager/) и [Medusa for Apache Cassandra](https://github.com/thelastpickle/cassandra-medusa).

Функции:

* Бэкап отдельного узла или целого кластера
  * экспорт схемы базы данных
  * снэпшоты всех или выбранных keyspaces
  * опциональное сжатие бэкапа с помощью `pigz` 
* Загрузка бэкапов в s3-совместимое хранилище через `awscli`
  * Автоматическое удаление бэкапов в хранилище после истечения заданного срока
* Обслуживание БД через вызов `nodetool repair`
* Поддержка вебхуков для отправки уведомлений о завершении работы и об ошибках

Планы:

* Восстановление из бэкапов
* Настройка github actions для сборки docker-образа и запуска тестов
---------------------

### Usage

* `scylla-octopus healthcheck` - performs a sanity check of the environment and configuration (scylladb status, the presence of required executables, etc)
* `scylla-octopus backup run` - runs a backup (exports database schema and snapshot, uploads to remote storage, cleans up)
* `scylla-octopus backup list` - prints a list of existing backups in remote storage
* `scylla-octopus backup list-expired` - prints a list of expired backups in remote storage that can be removed
* `scylla-octopus backup cleanup-expired` - removes expired backups from remote storage
* `scylla-octopus db list-snapshots` - prints a list of existing snapshots on database nodes
* `scylla-octopus db repair` - executes [nodetool repair -pr](https://docs.scylladb.com/operating-scylla/nodetool-commands/repair/) on database nodes

Command-line flags:

* `--config=...` - path to configuration file (defaults to `config/remote.yml`)
* `--verbose`, `-v` - forces debug output (equivalent to `log.level=debug` and `commands.debug=true` in configuration file)

### Configuration

See `config` directory for configuration examples.

`config/remote.yml` is an example for running a tool on multiple database nodes over SSH.
You will probably need to add a public SSH key to every machine beforehand.
In this mode, it doesn't matter where `scylla-octopus` is executed, as long as it can SSH to the nodes.

`config/local.yml` is an example for running a tool on a database node itself.
The options are mostly the same except the lack of `cluster.hosts` section.

### Requirements

`scylla-octopus` has a few assumptions about its environment:

* An [awscli](https://aws.amazon.com/cli/) executable should be available on every database node for backup uploading.
  * It can be used with any s3-compatible storage.
  * If it is unavailable, or you only want to keep local backups, then set `backup.disableUploading` to `true`.
  * An alternative storage implementation (such as `rsync`) would be welcomed.
* If backup compression is enabled with `archive.method: pigz`, then [pigz](https://zlib.net/pigz/) must be available on every database node.
  * So far `pigz` is the only supported compression method, but we're open to suggestions.
* Database nodes are running linux with an `sh` shell.
* The tool is tested with recent (4.x) scylladb versions, but will probably work with older ones too. 

### Error handling

A healthcheck is performed before backup and repair. If any node is unreachable, or has a status other than "UN" (up and running), the program stops.

Backups are executed in parallel on each node. If there is any error, then the misbehaving node is skipped, but the program doesn't stop.

Repairs are executed consecutively. If there is any error, the program stops and the remaining nodes will not be repaired.

### Building

If `go 1.17+` is installed locally, then `make build` will create an executable in `output/scylla-octopus`. 

Building docker image:

```
make docker-image

docker run --rm kolesa-team/scylla-octopus
```

### Development

For local development and testing we spin up 3 scylladb instances in docker-compose: `docker-compose up`.

Then, execute the following commands to configure the nodes (this will set up SSH keys on the nodes and install `awscli`):

```
make prepare-test-node node=scylla-node1
make prepare-test-node node=scylla-node2
make prepare-test-node node=scylla-node3

(this should really be automated with a single command)
``` 

You can also create a database (keyspace) with some testing data: `make init-db node=scylla-node1` (this will be replicated to every node).

Now try running a tool:

```
# healthcheck
go run main.go healthcheck

expected output (besides the debug logs):
{
  "10.5.0.2": "OK",
  "10.5.0.3": "OK",
  "10.5.0.4": "OK"
}

# backup
go run main.go backup run
```

See `test` directory for details about scylladb configuration in docker-compose.

#### Running on a database node directly

By default, a `config/remote.yml` configuration file is used to connect to database nodes over SSH.

Another execution mode is to run on database host directly.
This can also be tested with docker-compose. For each database container, `/scylla-octopus` directory is mounted with a program executable (compiled with `make build`).

Let's run `scylla-octopus` on node 1:

```
make build

docker-compose exec scylla-node1 sh

/scylla-octopus/scylla-octopus --config=/scylla/local.yml healthcheck

expected output:
{
  "10.5.0.2": "OK"
}
```

#### Inspecting backups

The easiest way to inspect backup contents is to set `backup.cleanupLocal` to `false`, run a backup, then SSH to a database host and navigate to `/var/lib/scylla/backup`:

```
# first make sure backup.cleanupLocal is false,
# then run a backup
go run main.go backup run

# SSH to database node
docker-compose exec scylla-node1 sh

# show the last backup metadata
cat /var/lib/scylla/backup/metadata.yml
```

---

© 2021 Kolesa Group. Licensed under [MIT](https://opensource.org/licenses/MIT)
