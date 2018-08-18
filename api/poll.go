package main

import "gopkg.in/mgo.v2/bson"

type poll struct {
	ID      bson.ObjectId  `bson:"_id" json:"id"`
	Title   string         `json:"title"`
	Options []string       `json:"options"`
	Result  map[string]int `json:"result,omitempty"`
}
