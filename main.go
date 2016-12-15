package main

import (
	"encoding/json"
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/robfig/cron"
	"github.com/streadway/amqp"
)

func listenPong(ch *amqp.Channel, msgs <-chan amqp.Delivery, db *mgo.Session) {
	for d := range msgs {
		LogCyan("Received a Pong:")
		log.Println(string(d.Body))
		mach := Machine{}
		json.Unmarshal(d.Body, &mach)
		machStr, _ := json.MarshalIndent(mach, "", "  ")
		LogMagenta("\n" + string(machStr))
		mach.Update(db)
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

// checkTimeout checks if any machine has timeout, and thus set its status to
// 'down'.
func checkTimeout(session *mgo.Session) {
	db := LoadDBName()
	c := session.DB(db).C("machine")
	iter := c.Find(bson.M{"status": "up"}).Iter()
	var machines []Machine
	iter.All(&machines)
	for _, m := range machines {
		pong := time.Time(m.Pong)
		d := time.Since(pong)
		if d.Seconds() > 60 {
			LogRed(m.Name + " is down!")
			m.SetDown(session)
		}
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
	// environment variables
	rabbitURI := LoadURI("rabbit")
	mongoURI := LoadURI("mongo")
	pingName := LoadRabbitName("ping")
	pongName := LoadRabbitName("pong")

	// connect to mongo
	dbSession, err := mgo.Dial(mongoURI)
	if err != nil {
		panic(err)
	}
	defer dbSession.Close()

	// connect to rabbit
	conn, ch := ConnectRabbit(rabbitURI)
	defer conn.Close()
	defer ch.Close()

	// braodcast via exchange per 60 seconds
	DeclareExchange(ch, pingName)
	c := cron.New()
	c.AddFunc("@every 60s", func() { ping(ch, pingName) })
	ping(ch, pingName)

	// listen to Pong from any responding machine
	DeclareQueue(ch, pongName)
	pongMsgs := ConsumeQueue(conn, ch, pongName, 1)
	go listenPong(ch, pongMsgs, dbSession)

	// check for any timeout from existing "up" machines
	c.AddFunc("@every 60s", func() { checkTimeout(dbSession) })
	checkTimeout(dbSession)

	// start cron
	c.Start()

	// listen by forever blocking
	forever := make(chan bool)
	// LogBlackOnWhite("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
