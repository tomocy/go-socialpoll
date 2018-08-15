package main

import (
	"log"

	nsqlib "github.com/bitly/go-nsq"
)

type nsq interface {
	publishVotes()
}

func newNSQ(addr string, voteCh <-chan string, closedCh chan<- struct{}) nsq {
	return newTwitterVoteNSQ(addr, voteCh, closedCh)
}

type twitterVoteNSQ struct {
	producer *nsqlib.Producer
	voteCh   <-chan string
	closedCh chan<- struct{}
}

func newTwitterVoteNSQ(addr string, voteCh <-chan string, closedCh chan<- struct{}) *twitterVoteNSQ {
	producer, _ := nsqlib.NewProducer(addr, nsqlib.NewConfig())
	return &twitterVoteNSQ{
		producer: producer,
		voteCh:   voteCh,
		closedCh: closedCh,
	}
}

func (nsq twitterVoteNSQ) publishVotes() {
	for vote := range nsq.voteCh {
		nsq.producer.Publish("vote", []byte(vote))
	}

	nsq.producer.Stop()
	log.Println("twitterVoteNSQ closed")
	nsq.closedCh <- struct{}{}
}
