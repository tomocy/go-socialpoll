package main

import (
	"log"

	nsq "github.com/bitly/go-nsq"
)

type twitterVoteNSQ struct {
	producer  *nsq.Producer
	stoppedCh chan struct{}
}

func newTwitterVoteNSQ(addr string) *twitterVoteNSQ {
	producer, _ := nsq.NewProducer(addr, nsq.NewConfig())
	return &twitterVoteNSQ{
		producer:  producer,
		stoppedCh: make(chan struct{}),
	}
}

func (nsq twitterVoteNSQ) publishVotes(votesCh <-chan string) {
	for vote := range votesCh {
		nsq.producer.Publish("vote", []byte(vote))
	}

	nsq.producer.Stop()
	log.Println("twitterVoteNSQ stopped publishing votes")
	nsq.notifyOfHavingStopped()
}

func (nsq twitterVoteNSQ) notifyOfHavingStopped() {
	nsq.stoppedCh <- struct{}{}
}
