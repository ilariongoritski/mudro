package main

import (
	"context"
	"log"

	"github.com/goritskimihail/mudro/services/rag-api/internal/http/ragapi"
)

func main() {
	if err := ragapi.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
