package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
)

type OrderRequest struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

type OrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type OrderInput struct {
	OrderID  string
	Item     string
	Quantity int
}

func processPayment(ctx context.Context) (string, error) {
	time.Sleep(100 * time.Millisecond)
	return "Payment processed", nil
}

func updateInventory(ctx context.Context) (string, error) {
	time.Sleep(50 * time.Millisecond)
	return "Inventory updated", nil
}

func sendNotification(ctx context.Context) (string, error) {
	time.Sleep(30 * time.Millisecond)
	return "Notification sent", nil
}

func orderWorkflow(ctx dbos.DBOSContext, input OrderInput) (string, error) {
	paymentResult, err := dbos.RunAsStep(ctx, processPayment)
	if err != nil {
		return "", err
	}

	inventoryResult, err := dbos.RunAsStep(ctx, updateInventory)
	if err != nil {
		return "", err
	}

	notificationResult, err := dbos.RunAsStep(ctx, sendNotification)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Order %s: %s for %s (qty: %d) | %s | %s | %s",
		input.OrderID, paymentResult, input.Item, input.Quantity,
		paymentResult, inventoryResult, notificationResult)
	return result, nil
}

func main() {
	dbURL := os.Getenv("DBOS_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/dbos?sslmode=disable"
	}

	ctx, err := dbos.NewDBOSContext(context.Background(), dbos.Config{
		DatabaseURL: dbURL,
		AppName:     "order-service",
	})
	if err != nil {
		log.Fatalf("Failed to create DBOS context: %v", err)
	}

	dbos.RegisterWorkflow(ctx, orderWorkflow)

	go func() {
		if err := dbos.Launch(ctx); err != nil {
			log.Fatalf("Failed to launch DBOS: %v", err)
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		orderID := fmt.Sprintf("ORD-%d", time.Now().Unix())

		input := OrderInput{
			OrderID:  orderID,
			Item:     req.Item,
			Quantity: req.Quantity,
		}

		handle, err := dbos.RunWorkflow(ctx, orderWorkflow, input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
			return
		}

		result, err := handle.GetResult()
		if err != nil {
			http.Error(w, fmt.Sprintf("Workflow failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := OrderResponse{
			OrderID: orderID,
			Status:  "completed",
			Message: result,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
