package main

import (
	"context"
	"os/exec"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)


func requestHandler(w http.ResponseWriter, r *http.Request) {
	logger := zap.Must(zap.NewDevelopment())
	defer logger.Sync()

	// Extract the request ID from the context or generate a new one if it doesn't exist
	requestID, ok := r.Context().Value("requestID").(string)
	if !ok {
		log.Printf("No Request ID in http request")
	}

	reqLogger := logger.With(
		zap.String("requestID", requestID),
	)

	// Define a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Command to be executed
	cmd := exec.CommandContext(ctx, "sleep", "10")

	// Run the command and capture its output
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if the error is due to context timeout
		if ctx.Err() == context.DeadlineExceeded {
			reqLogger.Error("Command timed out")
		} else {
			// Handle other errors
			reqLogger.Error("Command failed with error: %v\n", zap.Error(err))
		}
	}

	// Send response to the client
	fmt.Fprintf(w, "sleep: %s", output)
}


func main() {
	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// Handle requests with the requestHandler function and use context.WithValue to set the request ID in the context
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create a context with the request ID and pass it to the requestHandler
		ctx := context.WithValue(r.Context(), "requestID", uuid.New().String())
		requestHandler(w, r.WithContext(ctx))
	})

	// Start the HTTP server
	port := ":8080"
	log.Printf("Server listening on port %s", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
