package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"scaffold/manager/config"

	logger "github.com/jfcarter2358/go-logger"
	"github.com/streadway/amqp"
)

var runConn *amqp.Connection
var runChannel *amqp.Channel
var runQueue amqp.Queue

func handleError(err error, message string) {
	if err != nil {
		logger.Fatalf("", "Unexpected error: %s, %s", err.Error(), message)
	}
}

func RunManagerProducer() {
	var err error
	runConn, err = amqp.Dial(config.Config.RabbitMQConnectionString)
	handleError(err, "Can't connect to AMQP")

	runChannel, err = runConn.Channel()
	handleError(err, "Can't create a amqpChannel")

	runQueue, err = runChannel.QueueDeclare(config.Config.RunQueueName, true, false, false, false, nil)
	handleError(err, fmt.Sprintf("Could not declare %s queue", config.Config.RunQueueName))
}

func RunPublish(data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("", "Unable to marshal manager publish json: %s", err.Error())
		return err
	}

	err = runChannel.Publish("", runQueue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})

	if err != nil {
		logger.Errorf("", "Error publishing message: %s", err)
	}
	return err
}

func RunConsumer(receiveFunc func([]byte) error, queueName string) {
	conn, err := amqp.Dial(config.Config.RabbitMQConnectionString)
	handleError(err, "Can't connect to AMQP")
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare(queueName, true, false, false, false, nil)
	handleError(err, fmt.Sprintf("Could not declare worker queue %s", queueName))

	err = amqpChannel.Qos(1, 0, false)
	handleError(err, "Could not configure QoS")

	messageChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	handleError(err, "Could not register consumer")

	stopChan := make(chan bool)

	go func() {
		logger.Infof("", "Consumer ready, PID: %d", os.Getpid())
		for d := range messageChannel {
			logger.Tracef("", "Received a message: %s", d.Body)

			if err := receiveFunc(d.Body); err != nil {
				if err := d.Reject(true); err != nil {
					log.Printf("Error processing message : %s", err)
				} else {
					log.Printf("Nack-ed message")
				}
			}

			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			} else {
				log.Printf("Acknowledged message")
			}

		}
	}()

	// Stop for program termination
	<-stopChan
}
