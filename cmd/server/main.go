package main

import (
	"0219039_SistemasDistribuidos/server"
	"log"
)

// Create and start the server
func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
