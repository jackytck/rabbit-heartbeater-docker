package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// LoadURI loads the uri of rabbit server from .env.
func LoadURI() string {
	loadEnv()
	user := os.Getenv("RABBIT_USER")
	pwd := os.Getenv("RABBIT_PASSWORD")
	host := os.Getenv("RABBIT_HOST")
	port := os.Getenv("RABBIT_PORT")
	uri := fmt.Sprintf("amqp://%s:%s@%s:%s", user, pwd, host, port)
	return uri
}

// LoadRabbitQueue loads the target queue of rabbit server from .env.
func LoadRabbitName(name string) string {
	loadEnv()
	var queue string
	switch name {
	case "ping":
		queue = os.Getenv("RABBIT_EXCHANGE_ALTI_HEART_PING")
	case "pong":
		queue = os.Getenv("RABBIT_QUEUE_ALTI_HEART_PONG")
	}
	return queue
}

// LoadHostName loads the host name of this machine.
func LoadHostName() string {
	loadEnv()
	return os.Getenv("ALTI_HOST_NAME")
}
