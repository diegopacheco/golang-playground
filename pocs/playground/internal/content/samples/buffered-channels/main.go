package main

import "fmt"

func main() {
	queue := make(chan string, 3)
	for _, task := range []string{"build", "test", "deploy"} {
		queue <- task
		fmt.Printf("queued %-7s len=%d cap=%d\n", task, len(queue), cap(queue))
	}
	close(queue)
	for task := range queue {
		fmt.Println("processing", task)
	}
}
