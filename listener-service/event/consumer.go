package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	ampq "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
)

type Consumer struct {
	conn      *ampq.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *ampq.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}
	err := consumer.setup()
	if err != nil {
		log.Println("error setting up exchange")
		return consumer, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		log.Println("error creating channel", err)
		return err
	}

	return declareExchange(channel)
}

func (consumer *Consumer) Listen(topics []string) error {

	channel, err := consumer.conn.Channel()
	if err != nil {
		log.Println("error creating channel", err)
		return err
	}
	defer channel.Close()

	q, err := declareRandomQueue(channel)
	if err != nil {
		log.Println("error declaring queue	", err)
		return err
	}

	for _, s := range topics {
		channel.QueueBind(
			q.Name, // queue name
			s,
			"logs_topic",
			false,
			nil, //
		)

		if err != nil {
			log.Println("error binding queue", err)
			return err
		}
	}

	messages, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("error consuming messages", err)
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			log.Println("received message")
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("waiting for messages [Exchange, Queue]: [logs_topic, %s]\n", q.Name)
	<-forever
	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		log.Println("logging event siwtch")
		err := logEvent(payload)
		if err != nil {
			log.Println("error logging item", err)
		}
	case "auth":

	default:
		err := logEvent(payload)
		if err != nil {
			log.Println("error logging item", err)
		}
	}
}

func logEvent(p Payload) error {
	jsonData, _ := json.MarshalIndent(p, "", "\t")
	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	log.Println("logging event")

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}
