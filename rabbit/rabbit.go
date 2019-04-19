package rabbit

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

const EXCHANGE = "data"

type RMQConn struct {
	ch *amqp.Channel
	Ex string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func Connect(amqpURI string) (*RMQConn, error) {
	// Attempt to connect to RabbitMQ, Hopefully it is running
	// TODO: Create a retry ticker for the inital connection.
	conn, err := amqp.Dial(amqpURI)
	failOnError(err, "Failed to connect to RabbitMQ")
	log.Infof("Connected to RabbitMQ on  %s", amqpURI)

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		EXCHANGE, // name
		"topic",  // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // noWait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare the Exchange")
	log.Infof("Declaring/Connecting to RabbitMQ Exchange on  %s", EXCHANGE)

	return &RMQConn{ch, EXCHANGE}, nil
}

func (c *RMQConn) PublishMessage(m []byte) {

	err := c.ch.Publish(
		EXCHANGE, // exchange
		"a-key",  // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "application/json",
			Body:         m,
			Timestamp:    time.Now(),
		})

	failOnError(err, "Failed to Publish on RabbitMQ")

}
