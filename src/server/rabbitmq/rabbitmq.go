package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"scaffold/server/config"

	logger "github.com/jfcarter2358/go-logger"
	"github.com/streadway/amqp"
)

var killPublishConn *amqp.Connection
var killPublishChannel *amqp.Channel
var killPublishQueue amqp.Queue
var managerPublishConn *amqp.Connection
var managerPublishChannel *amqp.Channel
var managerPublishQueue amqp.Queue
var workerPublishConn *amqp.Connection
var workerPublishChannel *amqp.Channel
var workerPublishQueue amqp.Queue

func handleError(err error, message string) {
	if err != nil {
		logger.Fatalf("", "Unexpected error: %s, %s", err.Error(), message)
	}
}

func RunManagerProducer() {
	var err error
	managerPublishConn, err = amqp.Dial(config.Config.RabbitMQConnectionString)
	handleError(err, "Can't connect to AMQP")

	managerPublishChannel, err = managerPublishConn.Channel()
	handleError(err, "Can't create a amqpChannel")

	managerPublishQueue, err = managerPublishChannel.QueueDeclare(config.Config.WorkerQueueName, true, false, false, false, nil)
	handleError(err, fmt.Sprintf("Could not declare %s queue", config.Config.WorkerQueueName))
}

func ManagerPublish(data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("", "Unable to marshal manager publish json: %s", err.Error())
		return err
	}

	err = managerPublishChannel.Publish("", managerPublishQueue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})

	if err != nil {
		logger.Errorf("", "Error publishing message: %s", err)
	}
	return err
}

func RunWorkerProducer() {
	var err error
	workerPublishConn, err = amqp.Dial(config.Config.RabbitMQConnectionString)
	handleError(err, "Can't connect to AMQP")

	workerPublishChannel, err = workerPublishConn.Channel()
	handleError(err, "Can't create a amqpChannel")

	workerPublishQueue, err = workerPublishChannel.QueueDeclare(config.Config.ManagerQueueName, true, false, false, false, nil)
	handleError(err, fmt.Sprintf("Could not declare %s queue", config.Config.ManagerQueueName))
}

func WorkerPublish(data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("", "Unable to marshal worker publish json: %s", err.Error())
		return err
	}

	err = workerPublishChannel.Publish("", workerPublishQueue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})

	if err != nil {
		logger.Errorf("", "Error publishing message: %s", err)
	}
	return err
}

func RunKillProducer() {
	var err error
	killPublishConn, err = amqp.Dial(config.Config.RabbitMQConnectionString)
	handleError(err, "Can't connect to AMQP")

	killPublishChannel, err = killPublishConn.Channel()
	handleError(err, "Can't create a amqpChannel")

	killPublishQueue, err = killPublishChannel.QueueDeclare(config.Config.ManagerQueueName, true, false, false, false, nil)
	handleError(err, fmt.Sprintf("Could not declare %s queue", config.Config.ManagerQueueName))
}

func KillPublish(data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("", "Unable to marshal kill publish json: %s", err.Error())
		return err
	}

	err = killPublishChannel.ExchangeDeclare(
		config.Config.KillQueueName, // name
		"fanout",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		return err
	}

	err = killPublishChannel.Publish(config.Config.KillQueueName, "", false, false, amqp.Publishing{
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
