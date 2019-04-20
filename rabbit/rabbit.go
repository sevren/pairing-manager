package rabbit

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

const EXCHANGE = "data"

type RMQConn struct {
	ch *amqp.Channel
	Ex string
}

func warnOnError(err error, msg string) {
	log.Warnf("%s: %s", msg, err)
}

// Connect - Will attempt to set up a connection to the RabbitMQ server.
// if an error occurs in this process we will return it and disable
// rmq communication in this app.
func Connect(amqpURI string) (*RMQConn, error) {
	// Attempt to connect to RabbitMQ, Hopefully it is running
	// TODO: Create a retry ticker for the inital connection.
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		warnOnError(err, "Failed to connect to RabbitMQ")
		return nil, err
	}

	log.Infof("Connected to RabbitMQ on  %s", amqpURI)

	ch, err := conn.Channel()
	if err != nil {
		warnOnError(err, "Failed to open a channel")
		return nil, err
	}

	err = ch.ExchangeDeclare(
		EXCHANGE, // name
		"topic",  // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // noWait
		nil,      // arguments
	)
	if err != nil {
		warnOnError(err, "Failed to declare the Exchange")
		return nil, err
	}

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

	warnOnError(err, "Failed to Publish on RabbitMQ")

}
