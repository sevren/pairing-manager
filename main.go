package main

import (
	"flag"
	"net/http"

	"github.com/sevren/pair-man/rabbit"

	log "github.com/sirupsen/logrus"
)

var (
	amqpURI = flag.String("amqp", "amqp://guest:guest@localhost:5672/", "AMQP URI")
)

func main() {

	flag.Parse()

	// Attempt to connect to rabbitmq upon start up
	conn, err := rabbit.Connect(*amqpURI)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Setting up Routes for REST controller: localhost:8081")

	r, err := Routes(conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8081", r))

}
