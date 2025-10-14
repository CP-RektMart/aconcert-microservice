package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

var RabbitMQClient *RabbitMQ

// STRUCT TO STORE RABBITMQ CONNECTION AND CHANNEL
type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// FUNCTION TO CREATE A NEW RABBITMQ CONNECTION AND CHANNEL
func NewRabbitMQConnection(url string) {
	// CONNECT TO RABBITMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	// OPEN A RABBITMQ CHANNEL
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a RabbitMQ channel: %s", err)
	}

	// STORE RABBITMQ CONNECTION AND CHANNEL
	RabbitMQClient = &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}
}

// FUNCTION TO CONSUME A RABBITMQ CHANNEL
func (r *RabbitMQ) ConsumeRabbitMQQueue(queue_name string) (<-chan amqp.Delivery, error) {

	// DECLARE QUEUE NAME (IF NOT EXISTS)
	q, err := r.Channel.QueueDeclare(
		queue_name, // name of the queue
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)

	if err != nil {
		log.Printf("Failed to declare a RabbitMQ queue: %s", err)
		return nil, err
	}

	// SUBSCRIBE TO THE QUEUE
	msgs, err := r.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Printf("Failed to register a RabbitMQ consumer: %s", err)
		return nil, err
	}

	return msgs, nil
}

func (r *RabbitMQ) PublishToQueue(queue_name string, body []byte) error {
	// DECLARE QUEUE NAME (IF NOT EXISTS)
	_, err := r.Channel.QueueDeclare(
		queue_name, // name of the queue
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)

	if err != nil {
		log.Printf("Failed to declare a RabbitMQ queue: %s", err)
		return err
	}

	// PUBLISH MESSAGE TO QUEUE
	err = r.Channel.Publish(
		"",         // exchange
		queue_name, // routing key (queue name)
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Failed to publish a message to RabbitMQ: %s", err)
		return err
	}

	return nil
}

// FUNCTION TO CLOSE THE RABBITMQ CONNECTION AND CHANNEL
func (r *RabbitMQ) CloseConnection() {
	r.Channel.Close()
	r.Conn.Close()
}
