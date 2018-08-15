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
	stream            stream
	nsq               nsq
	interruptSignalCh chan os.Signal
	voteCh            chan string
	streamClosedCh    chan struct{}
	nsqClosedCh       chan struct{}
}

func newTwitterVote(dbURL string, nsqURL string) *twitterVote {
	interruptSignalCh := make(chan os.Signal)
	signal.Notify(interruptSignalCh, syscall.SIGINT)

	voteCh := make(chan string)
	streamClosedCh := make(chan struct{})
	nsqClosedCh := make(chan struct{})
	return &twitterVote{
		stream:            newStream(dbURL, voteCh, streamClosedCh),
		nsq:               newNSQ(nsqURL, voteCh, nsqClosedCh),
		interruptSignalCh: interruptSignalCh,
		voteCh:            voteCh,
		streamClosedCh:    streamClosedCh,
		nsqClosedCh:       nsqClosedCh,
	}
}

func (v *twitterVote) start() {
	log.Println("twitterVote started")
	go v.waitInterruptSignalToCloseStream()
	go v.closeConnectionToTwitterStreamPerSecond()

	go v.stream.open()

	go v.nsq.publishVotes()

	v.waitForStreamAndNSQToClose()
}

func (v *twitterVote) waitInterruptSignalToCloseStream() {
	log.Println("twitterVote waits interrupt signal to close stream")
	<-v.interruptSignalCh
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

func (v twitterVote) waitForStreamAndNSQToClose() {
	<-v.streamClosedCh
	close(v.voteCh)
	<-v.nsqClosedCh
}
