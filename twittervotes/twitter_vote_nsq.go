package main

import (
	"log"

	nsq "github.com/bitly/go-nsq"
)

type twitterVoteNSQ struct {
	producer *nsq.Producer
	closedCh chan<- struct{}
}

func newTwitterVoteNSQ(addr string, closedCh chan<- struct{}) *twitterVoteNSQ {
	producer, _ := nsq.NewProducer(addr, nsq.NewConfig())
	return &twitterVoteNSQ{
		producer: producer,
		closedCh: closedCh,
	}
}

func (nsq twitterVoteNSQ) publishVotes(voteCh <-chan string) {
	for vote := range voteCh {
		nsq.producer.Publish("vote", []byte(vote))
	}

	nsq.producer.Stop()
	log.Println("twitterVoteNSQ closed")
	nsq.closedCh <- struct{}{}
}
