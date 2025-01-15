package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"
)

func main() {

	rabbitConn, err := connect()
	if err != nil {
		log.Println("error connecting to rabbitmq", err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Println("error creating consumer", err)
		panic(err)
	}

	err = consumer.Listen([]string{"log.INFO", "log.ERROR", "log.WARNING"})
	if err != nil {
		log.Println("error listening to topics", err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("error connecting to rabbitmq", err)
			counts++
		} else {
			log.Println("connected to rabbitmq")
			connection = conn
			break
		}
		if counts > 5 {
			log.Println("error connecting to rabbitmq after 5 attempts")
			return nil, err
		}
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Backing off")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
