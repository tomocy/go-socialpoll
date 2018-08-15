package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type twitterVote struct {
	stream         *twitterStream
	nsq            *twitterVoteNSQ
	db             *twitterVoteDB
	termSignalCh   chan os.Signal
	voteCh         chan string
	streamClosedCh chan struct{}
	nsqClosedCh    chan struct{}
}

func newTwitterVote(dbURL string, nsqURL string) *twitterVote {
	termSignalCh := make(chan os.Signal)
	signal.Notify(termSignalCh, syscall.SIGINT)

	streamClosedCh := make(chan struct{})
	nsqClosedCh := make(chan struct{})
	return &twitterVote{
		stream:         newTwitterStream(streamClosedCh),
		nsq:            newTwitterVoteNSQ(nsqURL, nsqClosedCh),
		db:             newTwitterVoteDB(dbURL),
		termSignalCh:   termSignalCh,
		voteCh:         make(chan string),
		streamClosedCh: streamClosedCh,
		nsqClosedCh:    nsqClosedCh,
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

	v.waitForStreamAndNSQToClose()
}

func (v *twitterVote) waitInterruptSignalToCloseStream() {
	log.Println("twitterVote waits interrupt signal to close stream")
	<-v.termSignalCh
	fmt.Println()
	log.Println("twitterVote is closing stream")
	v.stream.close()
}

func (v *twitterVote) closeConnectionToTwitterStreamPerSecond() {
	log.Println("twitterVote closes connection to twitter per second")
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
		log.Fatalf("twitterVote could not dial to DB: %s\n", err)
	}
}

func (v twitterVote) closeDB() {
	v.db.close()
}

func (v twitterVote) publishVotes() {
	v.nsq.publishVotes(v.voteCh)
}

func (v twitterVote) openStream() {
	options, err := v.db.loadOptions()
	if err != nil {
		log.Fatalf("twitterVote could not load options from db: %s\n", err)
	}
	v.stream.start(v.voteCh, options)
}

func (v twitterVote) waitForStreamAndNSQToClose() {
	<-v.streamClosedCh
	close(v.voteCh)
	<-v.nsqClosedCh
}
