[![progress-banner](https://backend.codecrafters.io/progress/http-server/10eb84fc-8d9b-4c6f-afe1-24c6d1641fc0)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

# Simple HTTP Server 1.1

This is a [code crafters challenge](https://app.codecrafters.io/courses/http-server/overview) to build a simple HTTP server 1.1 in Go. Some implemented features are:

- Multiplexer
- Path variables
- Concurrent connections
- Persisten connections
- Gzip compression

To start the program:

```bash
$ go run main.go --directory /path/to/tmp # --directory is where files are written to and read from
```

To run tests:

```
$ make test
```

## Endpoints

```
GET /
Returns a 200 status code

GET /user-agent
Returns client user agent

GET /echo/{str}
Returns str

GET /files/{filename}
Read file from --directory

POST /files/{filename}
Create a new file with content from the request body in --directory
```

## Examples

Call the echo endpoint and gzip the response

```bash
$ curl -H "Accept-Encoding: gzip" http://localhost:4221/echo/hello | gunzip -c
```

Create a new file

```bash
$ curl --data "hello" -H "Content-Type: application/octet-stream" http://localhost:4221/files/hello
```

Fetch content of a file

```bash
$ curl http://localhost:4221/files/hello
```
