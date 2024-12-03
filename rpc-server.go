package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

// Define the service struct
type Arith struct{}

// Define a method for the Arith service
func (a *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

// Args struct holds the parameters for the function
type Args struct {
	A, B int
}

func main() {
	// Register the Arith service with the RPC server
	arith := new(Arith)
	err := rpc.Register(arith)
	if err != nil {
		log.Fatal("Error registering Arith service:", err)
	}

	// Start the RPC server in a separate goroutine
	go func() {
		listener, err := net.Listen("tcp", ":1234")
		if err != nil {
			log.Fatal("Error listening on port 1234:", err)
		}
		defer listener.Close()
		fmt.Println("RPC server is running on port 1234...")

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error accepting connection:", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	// HTTP handler for RPC through HTTP with CORS headers
	http.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			return
		}

		// Parse JSON input
		var args Args
		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Perform RPC call
		var result int
		err = arith.Multiply(&args, &result)
		if err != nil {
			http.Error(w, "RPC call failed", http.StatusInternalServerError)
			return
		}

		// Respond with the result
		response := map[string]int{"product": result}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start the HTTP server
	fmt.Println("HTTP server running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
