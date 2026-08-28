package main

import (
	"log"

	"github.com/CosmoS1X/go-project-278/internal/app"
)

func main() {
	router := app.NewRouter()

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
