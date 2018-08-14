package main

import "log"

func main() {
	twitterVote := newTwitterVote("mongodb://ballots:ballots@localhost/ballots", "localhost:4150")
	twitterVote.start()
	log.Println(11111111)
}
