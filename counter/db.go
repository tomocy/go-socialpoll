package main

import mgo "gopkg.in/mgo.v2"

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
