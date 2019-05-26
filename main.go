package main

import (
	"flag"
	"fmt"
	"log"

	"nmcapule/tiki/server"
)

var port = flag.Int("port", 1234, "port to use")

func main() {
	flag.Parse()

	log.Println("Listen to port:", *port)

	s := tiki.NewServer()
	if err := s.ListenAndServe(fmt.Sprintf(":%d", *port)); err != nil {
		panic(err)
	}
}
