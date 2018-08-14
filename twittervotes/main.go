package main

import (
	"log"

	nsq "github.com/bitly/go-nsq"
)

func main() {
	twitterStreamStopCh := make(chan struct{})
	twitterVote := newTwitterVote(twitterStreamStopCh)
	go twitterVote.waitInterruptSignalToFinishTwitterStream()
	go twitterVote.closeConnectionToTwitterStreamPerSecond()

	twitterVoteDB := initTwitterVoteDB("mongodb://ballots:ballots@localhost/ballots")
	if err := twitterVoteDB.dial(); err != nil {
		log.Fatalf("could not dial to DB: %s\n", err)
	}
	defer twitterVoteDB.close()

	votes := make(chan string)
	publisherStopCh := publishVotes(votes)
	twitterStreamStoppedCh := startTwitterStream(twitterStreamStopCh, votes, twitterVoteDB)

	<-twitterStreamStoppedCh
	close(votes)
	<-publisherStopCh
}

func publishVotes(votes <-chan string) <-chan struct{} {
	stop := make(chan struct{})
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	go func() {
		defer func() {
			stop <- struct{}{}
		}()

		for vote := range votes {
			pub.Publish("votes", []byte(vote))
		}

		log.Println("stopping publishing")
		pub.Stop()
		log.Println("stopped publishing")
	}()

	return stop
}
