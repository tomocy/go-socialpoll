package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	graceful "gopkg.in/tylerb/graceful.v1"
)

func main() {
	addr := flag.String("addr", ":8080", "the address of the endpoint")
	dbURL := flag.String("db-url", "localhost", "the url of the db")
	flag.Parse()

	log.Printf("connect to db: %s\n", *dbURL)
	dbSession, err := mgo.Dial(*dbURL)
	if err != nil {
		log.Fatalf("faild to connect to db: %s\n", err)
	}
	defer dbSession.Close()
	log.Println("successfully connected to db")

	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withVars(withDB(dbSession, withAPIKey(handlePolls)))))
	log.Println("start serving")
	graceful.Run(*addr, 1*time.Second, mux)
	log.Println("stopped serving")
}

func withCORS(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Header", "Location")
		f(w, r)
	}
}

func withVars(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		openVars(r)
		defer closeVars(r)

		f(w, r)
	}
}

func withDB(dbSession *mgo.Session, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ourDBSession := dbSession.Copy()
		defer ourDBSession.Close()

		setVar(r, "db", ourDBSession.DB("ballots"))
		f(w, r)
	}
}

func withAPIKey(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidAPIKey(r.URL.Query().Get("key")) {
			return
		}

		f(w, r)
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc"
}

func handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlePollsWithGet(w, r)
		return
	case "POST":
		handlePollsWithPost(w, r)
		return
	case "DELETE":
		handlePollsWithDelete(w, r)
		return
	}
	respondHTTPErr(w, http.StatusNotFound)
}

func handlePollsWithGet(w http.ResponseWriter, r *http.Request) {
	db := getVar(r, "db").(*mgo.Database)
	polls := db.C("polls")

	var q *mgo.Query
	path := newPath(r.URL.Path)
	if path.hasID() {
		q = polls.FindId(bson.ObjectIdHex(path.id))
	} else {
		q = polls.Find(nil)
	}

	var pollResults []*poll
	if err := q.All(&pollResults); err != nil {
		respondErr(w, http.StatusInternalServerError, err)
		return
	}

	respond(w, http.StatusOK, &pollResults)
}

func handlePollsWithPost(w http.ResponseWriter, r *http.Request) {
	db := getVar(r, "db").(*mgo.Database)
	polls := db.C("polls")

	var poll poll
	if err := decodeBody(r, &poll); err != nil {
		respondErr(w, http.StatusBadRequest, err)
		return
	}

	poll.ID = bson.NewObjectId()
	if err := polls.Insert(poll); err != nil {
		respondErr(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Location", "polls/"+poll.ID.Hex())
	respond(w, http.StatusCreated, nil)
}

func handlePollsWithDelete(w http.ResponseWriter, r *http.Request) {
	db := getVar(r, "db").(*mgo.Database)
	polls := db.C("polls")

	path := newPath(r.URL.Path)
	if !path.hasID() {
		respondErr(w, http.StatusMethodNotAllowed, "you cannot delete all polls")
		return
	}

	if err := polls.RemoveId(bson.ObjectIdHex(path.id)); err != nil {
		respondErr(w, http.StatusInternalServerError, err)
		return
	}

	respond(w, http.StatusOK, nil)
}
