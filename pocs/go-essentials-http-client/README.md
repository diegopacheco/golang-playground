# go-essentials-http-client

The `net/http` client: GET with query params, POST with a body, custom headers, response status/headers, and a client timeout. The server is a stdlib `httptest.Server` so it runs with no external service.

### How it works

`src/main.go` spins up an in-process `httptest.Server` with two routes, then drives it with an `http.Client` that has a timeout: a GET that reads a response header and body, and a POST built with `http.NewRequest` whose JSON response is streamed back with `json.NewDecoder`.

### Run

```bash
./run.sh
```

### Output

```
GET status: 200
GET body: hello alice
GET header X-Server: go-essentials
POST status: 200
POST decoded: map[received:payload]
```
