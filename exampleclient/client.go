package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var port = flag.Int("port", 1234, "port to connect into")

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(conn)
	for {
		fmt.Printf("")
		scanner.Scan()

		_, err := writer.WriteString(scanner.Text() + "\n")
		if err != nil {
			log.Println("error:", err)
		}
		writer.Flush()
	}
}
