package main

import (
	"log"
	"os"

	"todoapp/internal/app"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable required")
	}
	app.Init(secret)

	r := app.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
