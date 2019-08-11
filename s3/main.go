package main

import (
	"log"
	"net/http"
	"os"
)

func getPort() string {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func serveHTTP() {
	mountRoutes()
	port := getPort()

	log.Println("Starting server at :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	serveHTTP()
}
