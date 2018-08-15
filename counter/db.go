package main

import (
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type db interface {
	dial() error
	close()
	updateCounts(map[string]int)
}

func newDB(url string) db {
	return newMongoDB(url)
}

type mongoDB struct {
	url     string
	session *mgo.Session
}

func newMongoDB(url string) *mongoDB {
	return &mongoDB{
		url: url,
	}
}

func (db *mongoDB) dial() error {
	var err error
	db.session, err = mgo.Dial(db.url)
	return err
}

func (db *mongoDB) close() {
	db.session.Close()
}

func (db mongoDB) updateCounts(counts map[string]int) {
	polls := db.loadPolls()
	for option, count := range counts {
		selector := generateMongoDBSelector(option)
		updater := generateMongoDBUpdater(option, count)
		if _, err := polls.UpdateAll(selector, updater); err != nil {
			log.Printf("failed to update the count of %s: %s\n", option, err)
			continue
		}
	}
}

func (db mongoDB) loadPolls() *mgo.Collection {
	return db.session.DB("ballots").C("polls")
}

func generateMongoDBSelector(option string) bson.M {
	return bson.M{
		"options": bson.M{
			"$in": []string{option},
		},
	}
}

func generateMongoDBUpdater(option string, count int) bson.M {
	return bson.M{
		"$inc": bson.M{
			"results." + option: count,
		},
	}
}
