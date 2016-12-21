package main

import (
	"fmt"
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Machine represents one of the alti-worker.
type Machine struct {
	Name     string          `json:"name"`
	Nickname string          `json:"nickname"` // primary key
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
	c.Find(bson.M{"nickname": m.Nickname}).One(&old)
	if old.Status == "" || old.Status == "down" {
		msg := fmt.Sprintf("Host: %s\tType: %s\tNickname: %s\tIS UP!", m.Name, m.Type, m.Nickname)
		text := fmt.Sprintf("🚀 %s🚁%s🎈%s", m.Name, m.Type, m.Nickname)
		LogGreen(msg)
		SendTelegram(text)
		useHist = false
	}

	// append response to old
	if useHist {
		m.Response = append(m.Response, old.Response...)
	}
	m.Response = append(m.Response, d)
	m.Status = "up"

	// store the most recent response time only
	limit := LoadHistoryLimit()
	s := len(m.Response)
	if s > limit {
		m.Response = m.Response[s-limit:]
	}

	_, err := c.Upsert(bson.M{"nickname": m.Nickname}, m)
	if err != nil {
		log.Fatal(err)
	}
}

// SetDown sets the status to "down".
func (m *Machine) SetDown(session *mgo.Session) {
	db := LoadDBName()
	c := session.DB(db).C("machine")
	m.Status = "down"
	_, err := c.Upsert(bson.M{"nickname": m.Nickname}, m)
	if err != nil {
		log.Fatal(err)
	}
}
