package main

import (
	"log"
	"nosqli/internal/app"
)
func main() {
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}