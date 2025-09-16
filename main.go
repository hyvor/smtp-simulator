package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func getDomain() string {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}
	return domain
}

func main() {
	godotenv.Load()

	server := NewSmtpServer()

	log.Println("Starting server at", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
