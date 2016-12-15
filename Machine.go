package main

import (
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Machine represents one of the alti-worker.
type Machine struct {
	Name     string          `json:"name"`
	Nickname string          `json:"nickname"`
	Type     string          `json:"type"`
	Ping     JSONTime        `json:"ping"`
	Pong     JSONTime        `json:"pong"`
	Response []time.Duration `json:"response"`
	Extra    string          `json:"extra"`
	Status   string          `json:"status"`
}

// Update updates the struct in db.
func (m *Machine) Update(session *mgo.Session) {
	if m.Name == "" || m.Nickname == "" || m.Type == "" {
		return
	}
	db := LoadDBName()
	c := session.DB(db).C("machine")

	// calculate response time
	ping := time.Time(m.Ping)
	pong := time.Time(m.Pong)
	d := pong.Sub(ping)

	// get existing
	old := Machine{}
	useHist := true
	c.Find(bson.M{"name": m.Name}).One(&old)
	if old.Status == "" || old.Status == "down" {
		LogGreen(m.Name + " is up!")
		useHist = false
	}

	// append response to old
	if useHist {
		m.Response = append(m.Response, old.Response...)
	}
	m.Response = append(m.Response, d)
	m.Status = "up"

	_, err := c.Upsert(bson.M{"name": m.Name}, m)
	if err != nil {
		log.Fatal(err)
	}
}

// SetDown sets the status to "down".
func (m *Machine) SetDown(session *mgo.Session) {
	db := LoadDBName()
	c := session.DB(db).C("machine")
	m.Status = "down"
	_, err := c.Upsert(bson.M{"name": m.Name}, m)
	if err != nil {
		log.Fatal(err)
	}
}
