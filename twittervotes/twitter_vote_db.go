package main

import (
	"log"

	mgo "gopkg.in/mgo.v2"
)

type twitterVoteDB struct {
	url     string
	session *mgo.Session
}

func initTwitterVoteDB(url string) twitterVoteDB {
	return twitterVoteDB{
		url: url,
	}
}

func (db *twitterVoteDB) dial() error {
	log.Printf("twitter vote db dialed db: %s\n", db.url)
	var err error
	db.session, err = mgo.Dial(db.url)
	return err
}

func (db twitterVoteDB) close() {
	log.Printf("twitter vote db closed the connection to db: %s\n", db.url)
	db.session.Close()
}
