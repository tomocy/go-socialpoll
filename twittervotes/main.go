package main

import "log"

func main() {
	twitterStreamStopCh := make(chan struct{})
	twitterVote := newTwitterVote(twitterStreamStopCh, "mongodb://ballots:ballots@localhost/ballots", "localhost:4150")
	twitterVote.start()
	log.Println(11111111)
}
