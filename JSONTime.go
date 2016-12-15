package main

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// JSONTime represents a datetime.
type JSONTime time.Time

// MarshalJSON returns the momentjs like 'YYYY-MM-DD hh:mm A' formated datetime.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 03:04:05.0000 PM"))
	return []byte(stamp), nil
}

// UnmarshalJSON returns the parsed time object.
func (t *JSONTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*t = JSONTime(time.Now())
		return
	}
	loc, _ := time.LoadLocation("Asia/Hong_Kong")
	p, err := time.ParseInLocation("2006-01-02 03:04:05.0000 PM", s, loc)
	*t = JSONTime(p)
	return err
}

// GetBSON marshals JSONTime to time.Time.
func (t JSONTime) GetBSON() (interface{}, error) {
	return time.Time(t), nil
}

// SetBSON unmarshals back to JSONTime
func (t *JSONTime) SetBSON(raw bson.Raw) error {
	var tm time.Time
	if err := raw.Unmarshal(&tm); err != nil {
		return err
	}
	*t = JSONTime(tm)
	return nil
}
