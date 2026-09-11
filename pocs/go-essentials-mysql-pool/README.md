# go-essentials-mysql-pool

Connection pooling with `database/sql` and the MySQL driver: `SetMaxOpenConns`/`SetMaxIdleConns`/`SetConnMaxLifetime`, prepared statements, and pool stats under concurrent load. MySQL runs in podman.

### How it works

`docker-compose.yml` starts MySQL. `src/main.go` opens a pool, waits for the DB with a ping loop, creates a table, inserts through a prepared statement, reads the rows, fires 20 concurrent queries to exercise the pool, and prints `db.Stats()`.

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
pool: open=5 inUse=0 idle=5 maxOpen=10
done
```
