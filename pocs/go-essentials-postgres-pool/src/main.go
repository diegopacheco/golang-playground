package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connect(ctx context.Context) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig("postgres://app:app@127.0.0.1:5432/appdb")
	if err != nil {
		log.Fatal(err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return pool
}

func waitReady(ctx context.Context, pool *pgxpool.Pool) {
	for i := 0; i < 60; i++ {
		if err := pool.Ping(ctx); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	log.Fatal("database not ready")
}

func main() {
	ctx := context.Background()
	pool := connect(ctx)
	defer pool.Close()
	waitReady(ctx, pool)

	pool.Exec(ctx, `DROP TABLE IF EXISTS users`)
	pool.Exec(ctx, `CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT, created_at TIMESTAMP DEFAULT now())`)

	for _, name := range []string{"alice", "bob", "carol"} {
		pool.Exec(ctx, `INSERT INTO users (name) VALUES ($1)`, name)
	}

	var count int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	fmt.Println("rows inserted:", count)

	rows, _ := pool.Query(ctx, `SELECT id, name FROM users ORDER BY id`)
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Printf("user %d: %s\n", id, name)
	}
	rows.Close()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var n int
			pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
		}()
	}
	wg.Wait()

	stat := pool.Stat()
	fmt.Printf("pool: total=%d acquired=%d idle=%d maxConns=%d\n", stat.TotalConns(), stat.AcquiredConns(), stat.IdleConns(), stat.MaxConns())
	fmt.Println("done")
}
