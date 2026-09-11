package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

func connect() *sql.DB {
	mysql.SetLogger(log.New(io.Discard, "", 0))
	dsn := "app:app@tcp(127.0.0.1:3306)/appdb?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db
}

func waitReady(db *sql.DB) {
	for i := 0; i < 60; i++ {
		if err := db.Ping(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	log.Fatal("database not ready")
}

func main() {
	db := connect()
	defer db.Close()
	waitReady(db)

	db.Exec(`DROP TABLE IF EXISTS users`)
	db.Exec(`CREATE TABLE users (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(64), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)

	stmt, _ := db.Prepare(`INSERT INTO users (name) VALUES (?)`)
	defer stmt.Close()
	for _, name := range []string{"alice", "bob", "carol"} {
		stmt.Exec(name)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	fmt.Println("rows inserted:", count)

	rows, _ := db.Query(`SELECT id, name FROM users ORDER BY id`)
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
			db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
		}()
	}
	wg.Wait()

	stats := db.Stats()
	fmt.Printf("pool: open=%d inUse=%d idle=%d maxOpen=%d\n", stats.OpenConnections, stats.InUse, stats.Idle, stats.MaxOpenConnections)
	fmt.Println("done")
}
