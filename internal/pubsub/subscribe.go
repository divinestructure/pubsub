package pubsub

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T)) error {

	subCh, err := conn.Channel()
	if err != nil {
		return err
	}

	deliveries, err := subCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	errChan := make(chan error)

	go func() {
		for delivery := range deliveries {
			var message T

			if err := json.Unmarshal(delivery.Body, &message); err != nil {
				errChan <- err
				return
			}
			handler(message)

			if err := delivery.Ack(false); err != nil {
				errChan <- err
				return
			}

		}
	}()

	err = <-errChan

	return err
}
