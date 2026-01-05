package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](subCh *amqp.Channel, queueName, username string, handler func(T)) error {

	deliveries, err := subCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		var message T

		if err := json.Unmarshal(delivery.Body, &message); err != nil {
			return err
		}
		handler(message)

		if err := delivery.Ack(false); err != nil {
			return err
		}
		fmt.Println(string(delivery.Body))
		log.Printf("Move published successfully to %v.%v", routing.ArmyMovesPrefix, username)
	}

	return nil
}
