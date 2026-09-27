# go-playground

Hands-on Go POCs. Every project lives under [pocs/](pocs/), runs on its own and most ship with a `run.sh`.

## 🎮 Apps & Games

Full programs built end to end, not just a single feature.
Terminal games, dashboards and tools you can actually use.

* [tui-chess](pocs/tui-chess/) - Terminal chess with tcell, alpha-beta engine, perft-verified rules and neon sprites
* [tui-prod-manager-macos](pocs/tui-prod-manager-macos/) - Harbor: btop-style TUI for agent sessions, processes, Podman, cleanup and storage on macOS
* [brasileirao](pocs/brasileirao/) - Scrapes the Brazilian league table with rod and goquery, shows it in a tview TUI
* [docker-cleanup-cli](pocs/docker-cleanup-cli/) - Interactive tview CLI to search and delete Docker images
* [go-crash-course](pocs/go-crash-course/) - Idiomatic Go crash course, one file per topic

## 🎓 Go Essentials

A focused series on the parts of Go you use every day.
Each one is small, has no external services unless noted, and has its own README.

* [go-essentials-basics](pocs/go-essentials-basics/) - Loops, maps, switch, sort.Slice, structs, methods and embedding
* [go-essentials-idiomatic](pocs/go-essentials-idiomatic/) - Functional options, comma-ok, defer cleanup and zero values
* [go-essentials-hidden-gems](pocs/go-essentials-hidden-gems/) - Labeled break, iota, named returns, LIFO defer, maps/slices helpers
* [go-essentials-interfaces](pocs/go-essentials-interfaces/) - Implicit interfaces, type assertions, type switches and embedding
* [go-essentials-generics](pocs/go-essentials-generics/) - Custom constraints, cmp.Ordered, generic containers and comparable keys
* [go-essentials-functional](pocs/go-essentials-functional/) - Generic Map/Filter/Reduce, closures and composition
* [go-essentials-slices](pocs/go-essentials-slices/) - The slices package plus classic slice tricks and iterators
* [go-essentials-errors](pocs/go-essentials-errors/) - Sentinel errors, %w wrapping, errors.Is, errors.As and errors.Join
* [go-essentials-goroutines](pocs/go-essentials-goroutines/) - WaitGroup, Mutex, worker pool and sync.Once
* [go-essentials-channels](pocs/go-essentials-channels/) - Generator, pipeline, fan-in with select, close and time.After
* [go-essentials-semaphores](pocs/go-essentials-semaphores/) - Buffered-channel semaphore, RWMutex and atomic counters
* [go-essentials-context](pocs/go-essentials-context/) - Cancellation, timeouts, deadlines and request-scoped values
* [go-essentials-json](pocs/go-essentials-json/) - Struct tags, Marshal/Unmarshal, MarshalIndent, time.Time and streaming decode
* [go-essentials-rest](pocs/go-essentials-rest/) - REST API with only net/http and the Go 1.22 method+pattern router
* [go-essentials-http-client](pocs/go-essentials-http-client/) - net/http client: params, bodies, headers, timeouts against httptest
* [go-essentials-jwt](pocs/go-essentials-jwt/) - HS256 JWT with only the standard library
* [go-essentials-pii-aes](pocs/go-essentials-pii-aes/) - Encrypting PII with AES-256-GCM and detecting the wrong key
* [go-essentials-gorm](pocs/go-essentials-gorm/) - GORM on pure-Go in-memory SQLite: migrations, Preload and CRUD
* [go-essentials-mysql-pool](pocs/go-essentials-mysql-pool/) - database/sql pooling against MySQL in podman
* [go-essentials-postgres-pool](pocs/go-essentials-postgres-pool/) - pgxpool pooling against Postgres in podman

## 🧱 Language Basics

The building blocks of the language, one small program per concept.

* [basics](pocs/basics/) - First steps with Go
* [foundations](pocs/foundations/) - Methods, pointers and recover
* [arrays-fun](pocs/arrays-fun/) - Arrays
* [slices](pocs/slices/) - Slices with a sum and its test
* [for-fun](pocs/for-fun/) - Every form of for
* [constants-fun](pocs/constants-fun/) - Constants
* [values-fun](pocs/values-fun/) - Values and basic types
* [pointers-fun](pocs/pointers-fun/) - Pointers
* [closures-fun](pocs/closures-fun/) - Closures
* [recursion-fun](pocs/recursion-fun/) - Recursion
* [variadic-fun](pocs/variadic-fun/) - Variadic functions
* [multiple-returns-fun](pocs/multiple-returns-fun/) - Multiple return values
* [methodsfun](pocs/methodsfun/) - Methods on types
* [embeded-structures-fun](pocs/embeded-structures-fun/) - Struct embedding
* [interfaces-with-go](pocs/interfaces-with-go/) - Interfaces
* [stringsfun](pocs/stringsfun/) - The strings package
* [runes-fun](pocs/runes-fun/) - Runes and UTF-8
* [number-parsing-fun](pocs/number-parsing-fun/) - Parsing numbers from strings
* [time-fun](pocs/time-fun/) - Time and durations
* [line-filter-fun](pocs/line-filter-fun/) - A stdin line filter
* [go-bufio-char](pocs/go-bufio-char/) - Reading raw bytes with bufio
* [go-env-vars-simple](pocs/go-env-vars-simple/) - Environment variables
* [embedding-files-fun](pocs/embedding-files-fun/) - Embedding files with go:embed
* [ternary-if-hack](pocs/ternary-if-hack/) - An inline-if helper since Go has no ternary
* [simple-go-eclipse-project](pocs/simple-go-eclipse-project/) - Basic types, structs and random numbers
* [go-sample](pocs/go-sample/) - Package layout with a hello module and test
* [basic-algo](pocs/basic-algo/) - Stack and queue
* [GOMAXPROCS](pocs/GOMAXPROCS/) - How GOMAXPROCS limits CPU cores

## 🧬 Generics, Types & Reflection

Type parameters, aliases and looking at types at runtime.

* [go-1.18-generics-fun](pocs/go-1.18-generics-fun/) - Generics as they landed in Go 1.18
* [go-1.21-generics-fun](pocs/go-1.21-generics-fun/) - Generics with Go 1.21 additions
* [generic-fp-map-filter](pocs/generic-fp-map-filter/) - Generic map and filter
* [go-1-24-type-alias](pocs/go-1-24-type-alias/) - Generic type aliases from Go 1.24
* [reflect](pocs/reflect/) - reflect.Value, struct tags and setting fields
* [reflections](pocs/reflections/) - Reflection on structs and methods
* [godash-fun](pocs/godash-fun/) - godash: lodash-style helpers for Go

## 🧵 Concurrency

Goroutines and channels are what make Go different.
These cover the patterns and the traps.

* [go-routines](pocs/go-routines/) - Starting goroutines
* [goroutine-stateful-fun](pocs/goroutine-stateful-fun/) - State owned by a single goroutine
* [channel-buffering-fun](pocs/channel-buffering-fun/) - Buffered channels
* [channel-directions-fun](pocs/channel-directions-fun/) - Send-only and receive-only channels
* [channel-range-over-fun](pocs/channel-range-over-fun/) - Ranging over a channel
* [channel-select-fun](pocs/channel-select-fun/) - select over many channels
* [channel-timeouts-fun](pocs/channel-timeouts-fun/) - Timeouts with select
* [atomics-fun](pocs/atomics-fun/) - sync/atomic counters
* [safe-singleton](pocs/safe-singleton/) - Thread-safe singleton with sync.Once
* [synctest_fun](pocs/synctest_fun/) - testing/synctest: fake time for concurrent tests
* [go-signalcontext-fun](pocs/go-signalcontext-fun/) - Context cancelled by OS signals
* [graceful-shutdown](pocs/graceful-shutdown/) - Draining an HTTP server on SIGTERM
* [events-go](pocs/events-go/) - kelindar/event: typed in-process pub/sub
* [dbos-transact-fun](pocs/dbos-transact-fun/) - DBOS Transact: durable workflows on Go 1.25

## ⚠️ Errors & Panics

Go returns errors as values and uses panic only for the unexpected.

* [go-1.13-errors](pocs/go-1.13-errors/) - Error wrapping from Go 1.13
* [eris-fun](pocs/eris-fun/) - eris: errors with stack traces
* [panic-fun](pocs/panic-fun/) - panic
* [panic-recover](pocs/panic-recover/) - Recovering from a panic
* [recover-panic-fun](pocs/recover-panic-fun/) - recover inside defer
* [retry-fun](pocs/retry-fun/) - Retry with exponential backoff and context

## 🌐 HTTP & Web Frameworks

From plain net/http to full frameworks.
Most of them serve JSON on localhost.

* [minimalitic-rest-service](pocs/minimalitic-rest-service/) - REST with HTTP + JSON and no dependencies
* [web-fun](pocs/web-fun/) - A tiny net/http server
* [http](pocs/http/) - net/http server, mux and client
* [http-server-channels-goroutines](pocs/http-server-channels-goroutines/) - Handlers talking through channels
* [middleware-chain](pocs/middleware-chain/) - Logging and auth middleware chain
* [web](pocs/web/) - gorilla/mux routes and gorilla/websocket
* [go-rest-simple](pocs/go-rest-simple/) - REST with gorilla/mux
* [go-docker-rest](pocs/go-docker-rest/) - gorilla/mux REST service in a container
* [chi-fun](pocs/chi-fun/) - chi router
* [gin](pocs/gin/) - Gin
* [go-gin-gonic-sample](pocs/go-gin-gonic-sample/) - Gin REST service
* [fiber_fun](pocs/fiber_fun/) - Fiber
* [fast-http-fun](pocs/fast-http-fun/) - fasthttp
* [go-bee-simple](pocs/go-bee-simple/) - Beego with GoConvey tests
* [go-zero-fun](pocs/go-zero-fun/) - go-zero REST service
* [templ-fun](pocs/templ-fun/) - templ: type-safe HTML templates
* [templ-htmx](pocs/templ-htmx/) - templ + HTMX
* [req-v3-fun](pocs/req-v3-fun/) - req: a friendlier HTTP client
* [github-fetcher](pocs/github-fetcher/) - Pages through a GitHub org's repos with net/http
* [map-diff](pocs/map-diff/) - Diffs two lists of repos
* [machinebox-graphql-fun](pocs/machinebox-graphql-fun/) - GraphQL client with machinebox/graphql
* [mock-fun](pocs/mock-fun/) - dhuan/mock: mocking REST APIs
* [serverless-go-aws-lambda](pocs/serverless-go-aws-lambda/) - Go on AWS Lambda

## 📨 Messaging & Networking

Brokers, queues and raw sockets.

* [kafka](pocs/kafka/) - Kafka producer and consumer with confluent-kafka-go
* [nats-fun](pocs/nats-fun/) - NATS pub/sub
* [curlyq_fun](pocs/curlyq_fun/) - curlyq: background jobs on Redis
* [gws-fun](pocs/gws-fun/) - gws WebSocket server and client over KCP
* [evio](pocs/evio/) - evio: event-loop networking, a telnet server

## 🗄️ Databases & Storage

Clients for SQL, NoSQL and embedded key-value stores.

* [go-rds-mysql](pocs/go-rds-mysql/) - MySQL on RDS with database/sql
* [entc-fun](pocs/entc-fun/) - ent: entity framework from Facebook
* [go-cass-stress](pocs/go-cass-stress/) - Cassandra stress test with gocql
* [redis](pocs/redis/) - Redis with go-redis
* [dice-db-client-fun](pocs/dice-db-client-fun/) - DiceDB client
* [etcd-client-fun](pocs/etcd-client-fun/) - etcd v3 client
* [gorqlite-fun](pocs/gorqlite-fun/) - rqlite: distributed SQLite
* [title-38-client-fun](pocs/title-38-client-fun/) - Tile38 geospatial database client
* [bolt-fun](pocs/bolt-fun/) - BoltDB embedded key-value store
* [pebble-fun](pocs/pebble-fun/) - Pebble: CockroachDB's LSM key-value store
* [es](pocs/es/) - Elasticsearch with olivere/elastic

## 🔍 Search, Caching & Data Structures

In-memory structures that make lookups fast.

* [bleve-fun](pocs/bleve-fun/) - Bleve full-text search
* [roaring-bitmap-fun](pocs/roaring-bitmap-fun/) - Roaring bitmaps
* [imcache-fub](pocs/imcache-fub/) - imcache: generic in-memory cache
* [cache-fun](pocs/cache-fun/) - seaguest/cache: two-level cache in front of GORM

## 🔐 Security & Crypto

Hashing, encryption and password storage from the standard library and x/crypto.

* [aes-cipher-fun](pocs/aes-cipher-fun/) - AES encrypt and decrypt
* [bcrypt-fun](pocs/bcrypt-fun/) - Password hashing with bcrypt
* [sha-hash-fun](pocs/sha-hash-fun/) - SHA hashing

## 🖥️ TUIs, CLIs & Desktop

Terminal UIs, prompts, flags and desktop windows.

* [bubbletea-fun](pocs/bubbletea-fun/) - Bubble Tea with bubbles table and lipgloss
* [termui-fun](pocs/termui-fun/) - termui dashboard widgets
* [asciigraph-fun](pocs/asciigraph-fun/) - ASCII line charts in the terminal
* [mpb-fun](pocs/mpb-fun/) - mpb: multiple progress bars
* [tabby-fun](pocs/tabby-fun/) - tabby: aligned terminal tables
* [promptui-fun](pocs/promptui-fun/) - promptui prompts and selects
* [isshell-go-fun](pocs/isshell-go-fun/) - ishell: interactive shells
* [go-repl](pocs/go-repl/) - A REPL that evaluates expressions with govaluate
* [go-flags-fun](pocs/go-flags-fun/) - go-flags command-line parsing
* [glide-fun](pocs/glide-fun/) - cookoo CLI with the Glide package manager
* [key-press-go](pocs/key-press-go/) - Reading key presses with pkg/term
* [keypressed](pocs/keypressed/) - Reading key presses from raw stdin
* [fyne-fun](pocs/fyne-fun/) - Fyne desktop GUI

## 🧪 Testing & Benchmarks

Unit tests, benchmarks and load tests.

* [test-bench-fun](pocs/test-bench-fun/) - Tests and benchmarks with go test
* [benchs](pocs/benchs/) - String concat: +, bytes.Buffer and strings.Builder
* [go-time-bench](pocs/go-time-bench/) - Timing code by hand
* [assert-go-fun](pocs/assert-go-fun/) - assert-go assertions
* [dynamic-generic-tests-in-go](pocs/dynamic-generic-tests-in-go/) - Table tests generated with generics
* [k6-fun](pocs/k6-fun/) - k6 load test script

## 🧰 Config, Logging & Observability

Reading config, writing logs and sending metrics.

* [config-fun](pocs/config-fun/) - golobby/config with feeders
* [go-yaml-simple](pocs/go-yaml-simple/) - YAML with yaml.v2
* [log](pocs/log/) - The standard log package
* [go-log](pocs/go-log/) - phuslu/log: fast structured logging
* [slog-fun](pocs/slog-fun/) - log/slog structured logging
* [sfx](pocs/sfx/) - SignalFx datapoints and events

## 🧩 Parsers, Runtimes & Containers

Parsing languages, running JS and Wasm from Go, and driving containers.

* [participle-parser-fun](pocs/participle-parser-fun/) - participle: parser from Go struct tags
* [go-tree-sitter-fun](pocs/go-tree-sitter-fun/) - tree-sitter parsing JavaScript
* [v8go-fun](pocs/v8go-fun/) - Running JavaScript on V8
* [wasmer-go-fun](pocs/wasmer-go-fun/) - Running WebAssembly with Wasmer
* [containerd-fun](pocs/containerd-fun/) - Pulling and running Redis with the containerd client
* [geziyor-fun](pocs/geziyor-fun/) - Geziyor web scraping
