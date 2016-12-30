package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
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
		LogCyan("Done")
		d.Ack(false)
	}
}

// checkTimeout checks if any machine has timeout, and thus set its status to
// 'down'.
func checkTimeout(session *mgo.Session) {
	db := LoadDBName()
	limit := LoadTimeout()
	c := session.DB(db).C("machine")
	iter := c.Find(bson.M{"status": "up"}).Iter()
	var machines []Machine
	iter.All(&machines)

	var downs []Machine
	for _, m := range machines {
		pong := time.Time(m.Pong)
		d := time.Since(pong)
		if d.Seconds() > limit {
			// to prevent false negative, instead of doing in the loop
			// store all the downed machines for sending messages later
			downs = append(downs, m)
		}
	}

	// log and send down message
	for _, m := range downs {
		msg := fmt.Sprintf("Host: %s\tType: %s\tNickname: %s\tIS DOWN!", m.Name, m.Type, m.Nickname)
		text := fmt.Sprintf("💔 %s🔥%s💣%s", m.Name, m.Type, m.Nickname)
		LogRed(msg)
		SendTelegram(text)
		m.SetDown(session)
	}
}

// genStatusPage generates a static html status page.
func genStatusPage(session *mgo.Session) {
	// load all machines
	db := LoadDBName()
	c := session.DB(db).C("machine")
	iter := c.Find(bson.M{}).Sort("nickname").Iter()
	var machines []Machine
	iter.All(&machines)

	// page's data
	now := time.Now()
	p := Page{machines, now.Format("2006-01-02 03:04:05 PM MST")}

	// render template
	t, err := template.ParseFiles("status-template.html")
	if err != nil {
		LogRed(fmt.Sprintf("%v", err))
		return
	}

	// output to index.html
	f, err := os.Create("index.html")
	defer f.Close()
	if err != nil {
		LogRed(fmt.Sprintf("%v", err))
		return
	}
	err = t.Execute(f, p)
	if err != nil {
		LogRed(fmt.Sprintf("%v", err))
		return
	}

	// upload to s3
	UploadS3("index.html")
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

	// generate / update status page
	c.AddFunc("@every 60s", func() { genStatusPage(dbSession) })
	genStatusPage(dbSession)

	// start cron
	c.Start()

	// listen by forever blocking
	forever := make(chan bool)
	// LogBlackOnWhite("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
