package main

import (
	"context"
	"log"

	"github.com/goritskimihail/mudro/services/movie-catalog/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
