package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(state routing.PlayingState) {
		defer fmt.Print("> ")

		gs.HandlePause(state)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		fmt.Println("Move received:", move)

		defer fmt.Print("> ")

		gs.HandleMove(move)
	}
}

func main() {
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	fmt.Println("Starting Peril client...")
	const connectionString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("error establishing connection: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection established")

	pauseCh, pauseQueue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Printf("could not subscribe to pause: %v", err)
	} else {
		fmt.Printf("Pause-queue declared and bound %s\n", pauseQueue.Name)
	}

	moveCh, moveQueue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Printf("could not subscribe to move: %v", err)
	} else {
		fmt.Printf("Move-queue declared and bound %s\n", moveQueue.Name)
	}

	gs := gamelogic.NewGameState(username)

	go func() {
		if err := pubsub.SubscribeJSON(pauseCh, pauseQueue.Name, username, handlerPause(gs)); err != nil {
			log.Printf("SubscribeJSON to pause-queue error: %v", err)
			// } else {
		}
		fmt.Println("Successfully subscribed to Pause-queue")
	}()

	go func() {
		if err := pubsub.SubscribeJSON(moveCh, moveQueue.Name, username, handlerMove(gs)); err != nil {
			log.Printf("SubscribeJSON to move-queue error: %v", err)
			// } else {
		}
	}()
	fmt.Println("Successfully subscribed to Move-queue")

	for {
		words := gamelogic.GetInput()

		switch words[0] {
		case "spawn":
			if err := gs.CommandSpawn(words); err != nil {
				fmt.Println(err)
				continue
			}

		case "move":
			move, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("army has been moved")

			if err := pubsub.PublishJSON(moveCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, move); err != nil {
				log.Printf("Could not publish move: %v", err)
			} else {
				log.Printf("Move published successfully to %v.%v", routing.ArmyMovesPrefix, username)
			}

		case "status":
			gs.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("unknown command")
		}

	}

}
