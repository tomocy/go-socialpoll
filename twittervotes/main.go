package main

import (
	"log"
)

func main() {
	twitterVoteDB := initTwitterVoteDB("mongodb://ballots:ballots@localhost/ballots")
	if err := twitterVoteDB.dial(); err != nil {
		log.Fatalf("could not dial to DB: %s\n", err)
	}
	defer twitterVoteDB.close()

	twitterStreamStopCh := make(chan struct{})
	twitterVote := newTwitterVote(twitterVoteDB, twitterStreamStopCh)
	go twitterVote.waitInterruptSignalToFinishTwitterStream()
	go twitterVote.closeConnectionToTwitterStreamPerSecond()

	votes := make(chan string)
	twitterVoteNSQ := initTwitterVoteNSQ("localhost:4150")
	go twitterVoteNSQ.publishVotes(votes)

	go twitterVote.stream.start(twitterStreamStopCh, votes)

	<-twitterVote.stream.stoppedCh
	close(votes)
	<-twitterVoteNSQ.stoppedCh
}
