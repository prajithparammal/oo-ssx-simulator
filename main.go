package main

import (
	"log"
	"net/http"

	"github.dxc.com/terraform-providers/oo-ssx-simulator/ssx"
)

func main() {
	inMemStore := ssx.NewStore()
	server := ssx.NewServer(inMemStore)

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalf("could not listen on port 8080 %v", err)
	}
}
