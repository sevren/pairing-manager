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
	if err != nil {
		log.Warnf("%s: %s", msg, err)
	}
}

// Connect - Attempts to connect to rabbit mq server
// If it does not successfully connect then we will disable
// all rabbitmq functionality
func Connect(amqpURI string) (*RMQConn, error) {
	// Attempt to connect to RabbitMQ, Hopefully it is running
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Warn("Failed to connect to RabbitMQ")
		return nil, err
	}
	log.Infof("Connected to RabbitMQ on  %s", amqpURI)

	ch, err := conn.Channel()
	if err != nil {
		log.Warn("Failed to open a channel")
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
		log.Warn("Failed to declare the Exchange")
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
