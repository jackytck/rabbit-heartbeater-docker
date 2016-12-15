package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// LoadDBName loads the name of database in mongo.
func LoadDBName() string {
	loadEnv()
	return os.Getenv("MONGO_DB")
}

// LoadURI loads the uri of rabbit server from .env.
func LoadURI(key string) string {
	loadEnv()
	var uri string
	switch key {
	case "rabbit":
		user := os.Getenv("RABBIT_USER")
		pwd := os.Getenv("RABBIT_PASSWORD")
		host := os.Getenv("RABBIT_HOST")
		port := os.Getenv("RABBIT_PORT")
		uri = fmt.Sprintf("amqp://%s:%s@%s:%s", user, pwd, host, port)
	case "mongo":
		user := os.Getenv("MONGO_USER")
		pwd := os.Getenv("MONGO_PASSWORD")
		host := os.Getenv("MONGO_HOST")
		port := os.Getenv("MONGO_PORT")
		db := os.Getenv("MONGO_DB")
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", user, pwd, host, port, db)
	}
	return uri
}

// LoadRabbitName loads the target queue of rabbit server from .env.
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

// LoadTimeout loads the response timeout.
func LoadTimeout() float64 {
	loadEnv()
	ts := os.Getenv("RESPONSE_TIMEOUT")
	t, _ := strconv.Atoi(ts)
	return float64(t)
}

// LoadHistoryLimit loads the limit of the response history.
func LoadHistoryLimit() int {
	loadEnv()
	ls := os.Getenv("RESPONSE_HISTORY_LIMIT")
	l, _ := strconv.Atoi(ls)
	return l
}
