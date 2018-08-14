package main

import "log"

func main() {
	twitterStreamStopCh := make(chan struct{})
	twitterVote := newTwitterVote(twitterStreamStopCh, "mongodb://ballots:ballots@localhost/ballots", "localhost:4150")
	go twitterVote.waitInterruptSignalToFinishTwitterStream()
	go twitterVote.closeConnectionToTwitterStreamPerSecond()

	if err := twitterVote.stream.db.dial(); err != nil {
		log.Fatalf("could not dial to DB: %s\n", err)
	}
	defer twitterVote.stream.db.close()

	votes := make(chan string)
	go twitterVote.nsq.publishVotes(votes)

	go twitterVote.stream.start(twitterStreamStopCh, votes)

	<-twitterVote.stream.stoppedCh
	close(votes)
	<-twitterVote.nsq.stoppedCh
}
