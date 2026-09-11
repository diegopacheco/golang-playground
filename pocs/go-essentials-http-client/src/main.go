package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

func testServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Server", "go-essentials")
		fmt.Fprintf(w, "hello %s", r.URL.Query().Get("name"))
	})
	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"received": string(body)})
	})
	return httptest.NewServer(mux)
}

func main() {
	server := testServer()
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	resp, _ := client.Get(server.URL + "/hello?name=alice")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("GET status:", resp.StatusCode)
	fmt.Println("GET body:", string(body))
	fmt.Println("GET header X-Server:", resp.Header.Get("X-Server"))

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/echo", strings.NewReader("payload"))
	req.Header.Set("Content-Type", "text/plain")
	postResp, _ := client.Do(req)
	var decoded map[string]string
	json.NewDecoder(postResp.Body).Decode(&decoded)
	postResp.Body.Close()
	fmt.Println("POST status:", postResp.StatusCode)
	fmt.Println("POST decoded:", decoded)
}
