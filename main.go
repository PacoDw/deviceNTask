package main

import (
	"log"

	"github.com/PacoDw/deviceNTask/dnt"
)

func main() {
	if err := dnt.CreateOptimalConfigurationFile("challenge.in"); err != nil {
		log.Fatal(err)
	}
}
