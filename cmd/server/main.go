package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const connectionString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("error establishing connection: %v", err)
	}
	defer conn.Close()
	fmt.Println("Successful connection to RabbitMQ")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			log.Println("sending a pause message")
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			}); err != nil {
				log.Fatalf("couldn't publish message %v", err)
			}

		case "resume":
			log.Println("sending a resume message")
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			}); err != nil {
				log.Fatalf("couldn't publish message %v", err)
			}

		case "quit":
			log.Println("exiting")
			return

		default:
			log.Println("unknown command")

		}

	}

}
