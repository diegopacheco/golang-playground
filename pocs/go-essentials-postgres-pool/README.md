# go-essentials-postgres-pool

Connection pooling with `pgx/v5` `pgxpool`: `MaxConns`/`MinConns`/`MaxConnLifetime`, parameterized queries, and pool stats under concurrent load. Postgres runs in podman.

### How it works

`docker-compose.yml` starts Postgres. `src/main.go` builds a `pgxpool.Pool` from a parsed config, waits for the DB with a ping loop, creates a table, inserts with `$1` placeholders, reads the rows, fires 20 concurrent queries to exercise the pool, and prints `pool.Stat()`.

### Run

```bash
./start.sh
./run.sh
./stop.sh
```

Or the whole cycle at once:

```bash
./test.sh
```

### Output

```
rows inserted: 3
user 1: alice
user 2: bob
user 3: carol
pool: total=10 acquired=0 idle=10 maxConns=10
done
```
