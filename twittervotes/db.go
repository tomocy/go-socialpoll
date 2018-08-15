package main

import (
	"log"

	mgo "gopkg.in/mgo.v2"
)

type db interface {
	dial() error
	close()
	loadOptions() ([]string, error)
}

func newDB(url string) db {
	return newTwitterVoteDB(url)
}

type twitterVoteDB struct {
	url     string
	session *mgo.Session
}

func newTwitterVoteDB(url string) *twitterVoteDB {
	return &twitterVoteDB{
		url: url,
	}
}

func (db *twitterVoteDB) dial() error {
	var err error
	db.session, err = mgo.Dial(db.url)
	log.Printf("twitterVoteDB dialed db: %s\n", db.url)
	return err
}

func (db *twitterVoteDB) close() {
	db.session.Close()
	log.Printf("twitterVoteDB closed the connection to db: %s\n", db.url)
}

type poll struct {
	Options []string
}

func (db twitterVoteDB) loadOptions() ([]string, error) {
	options := make([]string, 0)
	var poll poll
	pollsIter := db.session.DB("ballots").C("polls").Find(nil).Iter()
	for pollsIter.Next(&poll) {
		options = append(options, poll.Options...)
	}

	return options, pollsIter.Err()
}
