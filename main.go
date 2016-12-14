package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/robfig/cron"
	"github.com/streadway/amqp"
)

func listenPong(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		LogCyan("Received a Pong:")
		log.Println(string(d.Body))
		mach := Machine{}
		json.Unmarshal(d.Body, &mach)
		machStr, _ := json.MarshalIndent(mach, "", "  ")
		LogMagenta("\n" + string(machStr))
		// // set start time
		// proj.DateStartArchive = JSONTime(time.Now())
		// err := archive(dataPath, proj.PID)
		// if onError(err, d, proj, "Failed to archive") {
		// 	continue
		// }
		// // set ssh host name
		// proj.SourceHost = host
		// // set absolute path of archive
		// file := fmt.Sprintf("%s/%s.tar", depart, proj.PID)
		// file, _ = filepath.Abs(file)
		// proj.SourceFile = file
		// // set filesize of archive
		// size, err := getArchiveSize(file)
		// if onError(err, d, proj, "Failed to get filesize of archive") {
		// 	continue
		// }
		// proj.Size = size
		// // set end time
		// proj.DateEndArchive = JSONTime(time.Now())
		// // requeue to rabbit for alti-transferer to process
		// data, err := json.Marshal(proj)
		// if onError(err, d, proj, "Failed to marshal json") {
		// 	continue
		// }
		// err = Publish(ch, transfer, data)
		// if onError(err, d, proj, "Failed to publish to "+transfer) {
		// 	continue
		// }
		LogCyan("Done")
		d.Ack(false)
	}
}

func ping(ch *amqp.Channel, ex string) {
	LogCyan("Sending Ping...")
	machine := Machine{}
	machine.Ping = JSONTime(time.Now())
	data, err := json.Marshal(machine)
	FailOnError(err, "Failed to marshal machine json")
	PublishExchange(ch, ex, data)
	LogCyan("Done")
}

func main() {
	uri := LoadURI()
	pingName := LoadRabbitName("ping")
	pongName := LoadRabbitName("pong")

	conn, ch := ConnectRabbit(uri)
	defer conn.Close()
	defer ch.Close()

	DeclareExchange(ch, pingName)
	// broadcast to all machines per 60 seconds
	c := cron.New()
	c.AddFunc("@every 60s", func() { ping(ch, pingName) })
	c.Start()

	// listen to Pong from any responding machine
	DeclareQueue(ch, pongName)
	pongMsgs := ConsumeQueue(conn, ch, pongName, 1)
	go listenPong(ch, pongMsgs)

	forever := make(chan bool)
	// LogBlackOnWhite("[*] Waiting for messages. To exit press CTRL+C")

	<-forever
}
