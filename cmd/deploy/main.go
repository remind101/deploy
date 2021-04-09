package main

import (
	"log"
	"os"

	"deploy"
)

func main() {
	app := deploy.NewApp()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
