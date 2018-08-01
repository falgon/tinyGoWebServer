package main

import (
	"flag"
	"log"
	"./sync"
)

func init() {
	flag.Parse()
}

func main() {
	if app, err := sync.NewAppServer("tcp"); err != nil {
		log.Fatal(err)
	} else {
		app.Serve()
	}
}
