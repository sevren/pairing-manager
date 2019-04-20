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
	// If unsucessful we will warn but still start the service.. so you can use the REST controller at least.
	// Please refer to the documentation for Running this micro service
	conn, err := rabbit.Connect(*amqpURI)
	if err != nil {
		log.Warn("Could not connect to RabbitMQ server, (Challenge 3) RMQ Publishing disabled!\n", err)
	}

	log.Info("Setting up Routes for REST controller: localhost:8081")

	r, err := Routes(conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8081", r))

}
