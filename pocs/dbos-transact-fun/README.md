## DBOS Transact Golang POC

Go 1.25 proof of concept using DBOS Transact for durable workflow orchestration.

### Start

```bash
./start.sh
```

### Test

```bash
./test.sh
```

### Stop

```bash
./stop.sh
```

### In this POC

- Durable workflow orchestration with PostgreSQL
- Order processing workflow with payment, inventory, and notification steps
- HTTP API endpoints
- Docker Compose setup with Podman support

### Results

```bash
❯ ./start.sh
WARN[0000] Docker Compose is configured to build using Bake, but buildx isn't installed
[+] Building 4.5s (17/17) FINISHED                                                                  docker-container:default
 => [app internal] load build definition from Dockerfile                                                                0.0s
 => => transferring dockerfile: 252B                                                                                    0.0s
 => [app internal] load metadata for docker.io/library/alpine:latest                                                    0.7s
 => [app internal] load metadata for docker.io/library/golang:1.25-alpine                                               0.7s
 => [app internal] load .dockerignore                                                                                   0.0s
 => => transferring context: 115B                                                                                       0.0s
 => [app builder 1/6] FROM docker.io/library/golang:1.25-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f5  0.0s
 => => resolve docker.io/library/golang:1.25-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c0  0.0s
 => [app internal] load build context                                                                                   0.0s
 => => transferring context: 784B                                                                                       0.0s
 => [app stage-1 1/3] FROM docker.io/library/alpine:latest@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6  0.0s
 => => resolve docker.io/library/alpine:latest@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1  0.0s
 => CACHED [app builder 2/6] WORKDIR /app                                                                               0.0s
 => CACHED [app builder 3/6] COPY go.mod go.sum* ./                                                                     0.0s
 => CACHED [app builder 4/6] RUN go mod download                                                                        0.0s
 => [app builder 5/6] COPY . .                                                                                          0.0s
 => [app builder 6/6] RUN go build -o main .                                                                            3.6s
 => CACHED [app stage-1 2/3] WORKDIR /app                                                                               0.0s
 => CACHED [app stage-1 3/3] COPY --from=builder /app/main .                                                            0.0s
 => [app] exporting to docker image format                                                                              0.2s
 => => exporting layers                                                                                                 0.0s
 => => exporting manifest sha256:927f53fee3e0a0b6aeb4a162e2862acf6475616a686ed7f3cab0a0265fcca1c9                       0.0s
 => => exporting config sha256:4eb056a172cf1b03706cbe0b7eab32161d2015fbbcee48b33d4e9c26904dd18e                         0.0s
 => => sending tarball                                                                                                  0.2s
 => [app] importing to docker                                                                                           0.0s
 => [app] resolving provenance for metadata file                                                                        0.0s
[+] Building 1/1
 ✔ app  Built                                                                                                           0.0s
[+] Running 4/4
 ✔ Network dbos-transact-fun_default         Created                                                                    0.0s
 ✔ Volume "dbos-transact-fun_postgres-data"  Created                                                                    0.0s
 ✔ Container dbos-postgres                   Healthy                                                                    1.6s
 ✔ Container dbos-app                        Started                                                                    1.6s
Waiting for services to be healthy...
PostgreSQL is ready
Application is ready
All services are running
NAME            IMAGE                                            COMMAND      SERVICE    CREATED         STATUS                  PORTS
dbos-app        docker.io/library/dbos-transact-fun-app:latest   "./main"     app        2 seconds ago   Up Less than a second   0.0.0.0:8080->8080/tcp
dbos-postgres   docker.io/library/postgres:16-alpine             "postgres"   postgres   2 seconds ago   Up 1 second             0.0.0.0:5432->5432/tcp
❯ ls
docker-compose.yml go.mod             main.go            run.sh             stop.sh
Dockerfile         go.sum             README.md          start.sh           test.sh
❯ ./test.sh
Testing health endpoint...
Health response: {"status":"healthy"}
Health check passed

Testing order endpoint...
Order response: {"order_id":"ORD-1759734297","status":"completed","message":"Order ORD-1759734297: Payment processed for laptop (qty: 2) | Payment processed | Inventory updated | Notification sent"}
Order test passed

Testing another order...
Second order response: {"order_id":"ORD-1759734297","status":"completed","message":"Order ORD-1759734297: Payment processed for phone (qty: 5) | Payment processed | Inventory updated | Notification sent"}
Second order test passed
```