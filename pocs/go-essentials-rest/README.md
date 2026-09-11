# go-essentials-rest

A REST API with only `net/http`: the Go 1.22+ method+pattern router (`GET /tasks`, `POST /tasks`, `GET /tasks/{id}`), path values, JSON encoding, and a mutex-guarded in-memory store.

### How it works

`src/main.go` registers routes on an `http.ServeMux` using method-prefixed patterns and `r.PathValue("id")`, backs them with a `sync.Mutex`-protected map, and returns JSON with the right status codes (`201 Created`, `404 Not Found`). No web framework.

### Run

```bash
./run.sh
```

The server listens on `:8080`. Drive it from another terminal:

```bash
curl -X POST localhost:8080/tasks -d '{"title":"learn go","done":false}'
curl localhost:8080/tasks
curl localhost:8080/tasks/1
```

### Output

```
POST /tasks -> 201 {"id":1,"title":"learn go","done":false}
POST /tasks -> 201 {"id":2,"title":"write tests","done":true}
GET /tasks -> 200 [{"id":1,"title":"learn go","done":false},{"id":2,"title":"write tests","done":true}]
GET /tasks/1 -> 200 {"id":1,"title":"learn go","done":false}
GET /tasks/99 -> 404 {"error":"not found"}
```
