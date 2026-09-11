package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type User struct {
	Name   string   `json:"name"`
	Age    int      `json:"age"`
	Roles  []string `json:"roles"`
	Active bool     `json:"active"`
}

type Event struct {
	Name string    `json:"name"`
	At   time.Time `json:"at"`
}

func marshal() string {
	u := User{Name: "alice", Age: 30, Roles: []string{"admin", "user"}, Active: true}
	raw, _ := json.Marshal(u)
	return string(raw)
}

func unmarshal() User {
	var u User
	json.Unmarshal([]byte(`{"name":"bob","age":25,"roles":["user"],"active":false}`), &u)
	return u
}

func prettyPrint() string {
	data := map[string]any{"a": 1, "b": 2, "nested": map[string]any{"x": 10, "y": 20}}
	raw, _ := json.MarshalIndent(data, "", "  ")
	return string(raw)
}

func customTime() string {
	e := Event{Name: "login", At: time.Date(2026, 7, 12, 9, 30, 0, 0, time.UTC)}
	raw, _ := json.Marshal(e)
	return string(raw)
}

func stream() []string {
	dec := json.NewDecoder(strings.NewReader(`{"id":1}{"id":2}{"id":3}`))
	out := []string{}
	for dec.More() {
		var m map[string]int
		dec.Decode(&m)
		out = append(out, fmt.Sprintf("id=%d", m["id"]))
	}
	return out
}

func main() {
	fmt.Println("marshal:", marshal())
	fmt.Printf("unmarshal: %+v\n", unmarshal())
	fmt.Println("prettyPrint:")
	fmt.Println(prettyPrint())
	fmt.Println("customTime:", customTime())
	fmt.Println("stream:", stream())
}
