package main

import (
	"log"

	nsqlib "github.com/bitly/go-nsq"
)

type nsq interface {
	publishVotes(<-chan string)
}

type twitterVoteNSQ struct {
	producer *nsqlib.Producer
	closedCh chan<- struct{}
}

func newTwitterVoteNSQ(addr string, closedCh chan<- struct{}) *twitterVoteNSQ {
	producer, _ := nsqlib.NewProducer(addr, nsqlib.NewConfig())
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
