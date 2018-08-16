package main

import (
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

func main() {

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

func withDB(dbSession *mgo.Session, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ourDBSession := dbSession.Copy()
		defer ourDBSession.Close()

		setVar(r, "db", ourDBSession.DB("ballots"))
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

func withCORS(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Header", "Location")
		f(w, r)
	}
}
