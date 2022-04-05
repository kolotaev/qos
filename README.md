# QoS Example

### Build status

[![Unit Tests](https://github.com/kolotaev/qos/workflows/Unit%20tests/badge.svg?branch=master)](https://github.com/kolotaev/qos/actions)

[![E2E Tests](https://github.com/kolotaev/qos/workflows/E2E%20tests/badge.svg?branch=master)](https://github.com/kolotaev/qos/actions)


### Description

This repository is an example of a bandwidth limiting for TCP file servers.

There are 2 example File servers that are serving files from the base directory and 1 Administration server that allows to configure bandwidth limits for servers and individual connections using TCP text commands interface.

File server uses a `Throttler` to serve files to individual clients. Throttler uses 1 second resolution and allows to
set bandwidth limits in bytes. Thus minimum bandwidth value is `1 b/s` which is a fair minimum for a practical usage. Limits can be set for the whole server (applies to all existing connections) and for individual connection by connection's remote address (`host:port`). Limits can be set in runtime via an Administration server (through a TCP text interface) and are immediately applied to existing and new connections.


### How to run:

- Clone the repository.
- `cd /path/to/repo`.
- Check contents of `example/main.go` and adjust needed values.
- `make run`
- In a separate terminal window run `nc 127.0.0.1 3000` thus establishing connection with the 1st File server.
- In a separate terminal window run `nc 127.0.0.1 5000` thus establishing connection with Administration server.
- Type in commands from the list below.
- Observe the results.


### List of commands:

All the commands are modeled after most common TCP tools (e.g. Redis).

All limits are in bytes.

| Command | Admin (A) or File (F) server? | Description | 
| ------ | ----------- | ----- |
| STOP   | A, F | Stop server. |
| FILE | F | Download a file (args: file_name). |
| THROTTLE    | A | Enable or disable throttling for a server (args: yes/no). |
| SLIMIT    | A | Set bandwidth limit per server (args: srv_name limit_number). |
| CLIMIT    | A | Set bandwidth limit per connection (args: srv_name limit_number). |

Examples:

- `THROTTLE srv1 yes`
- `THROTTLE srv1 no`
- `SLIMIT srv2 35`
- `CLIMIT 127.0.0.1:51637 50`


### How to test:

- Clone the repository.
- `cd /path/to/repo`.
- `make test-unit` # to run unit tests
- `make test-unit` # to run unit tests
- `make test` # to run all tests
