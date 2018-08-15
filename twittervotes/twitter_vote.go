package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type twitterVote struct {
	stream       *twitterStream
	nsq          *twitterVoteNSQ
	db           *twitterVoteDB
	termSignalCh chan os.Signal
	votesCh      chan string
}

func newTwitterVote(dbURL string, nsqURL string) *twitterVote {
	termSignalCh := make(chan os.Signal)
	signal.Notify(termSignalCh, syscall.SIGINT)
	return &twitterVote{
		stream:       newTwitterStream(),
		nsq:          newTwitterVoteNSQ(nsqURL),
		db:           newTwitterVoteDB(dbURL),
		termSignalCh: termSignalCh,
		votesCh:      make(chan string),
	}
}

func (v *twitterVote) start() {
	log.Println("twitterVote started")
	go v.waitInterruptSignalToCloseStream()
	go v.closeConnectionToTwitterStreamPerSecond()

	v.dialDB()
	defer v.closeDB()

	go v.publishVotes()
	go v.openStream()

	v.waitForStreamAndNSQToBeClosed()
}

func (v *twitterVote) waitInterruptSignalToCloseStream() {
	log.Println("twitterVote is waiting interrupt signal to finish twitter stream")
	<-v.termSignalCh
	log.Println("twitterVote is stopping and finishing twitter stream")
	v.stream.close()
}

func (v *twitterVote) closeConnectionToTwitterStreamPerSecond() {
	log.Println("twitterVote closes connection to twitter stream per second")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			v.stream.closeConnection()
		}
	}
}

func (v twitterVote) dialDB() {
	if err := v.db.dial(); err != nil {
		log.Fatalf("could not dial to DB: %s\n", err)
	}
}

func (v twitterVote) closeDB() {
	v.db.close()
}

func (v twitterVote) publishVotes() {
	v.nsq.publishVotes(v.votesCh)
}

func (v twitterVote) openStream() {
	options, err := v.db.loadOptions()
	if err != nil {
		log.Fatalf("could not load options from db: %s\n", err)
	}
	v.stream.start(v.votesCh, options)
}

func (v twitterVote) waitForStreamAndNSQToBeClosed() {
	<-v.stream.closedCh
	close(v.votesCh)
	<-v.nsq.closedCh
}
