package main

import (
	"log"
)

func main() {
	server := NewSmtpServer()

	log.Println("Starting server at", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
