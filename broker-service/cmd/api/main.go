package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const webPort = "8080"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {

	rabbitConn, err := connect()
	if err != nil {
		log.Println("error connecting to rabbitmq", err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}
	log.Printf("starting broker service on port %s", webPort)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Println("server started")

	err = server.ListenAndServe()
	if err != nil {
		log.Panic(err)
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
