package main

import nsqlib "github.com/bitly/go-nsq"

type messageQueue interface {
	stop()
	connectToLookupd(string) error
}

func newMessageQueue(incrementCh chan<- string) (messageQueue, error) {
	return newNSQ(incrementCh)
}

type nsq struct {
	consumer    *nsqlib.Consumer
	incrementCh chan<- string
}

func newNSQ(incrementCh chan<- string) (*nsq, error) {
	consumer, err := newNSQConsumer(incrementCh)
	if err != nil {
		return nil, err
	}

	return &nsq{
		consumer:    consumer,
		incrementCh: incrementCh,
	}, nil
}

func newNSQConsumer(incrementCh chan<- string) (*nsqlib.Consumer, error) {
	consumer, err := nsqlib.NewConsumer("votes", "counter", nsqlib.NewConfig())
	if err != nil {
		return nil, err
	}
	consumer.AddHandler(nsqlib.HandlerFunc(func(m *nsqlib.Message) error {
		option := string(m.Body)
		incrementCh <- option
		return nil
	}))

	return consumer, nil
}

func (nsq nsq) stop() {
	nsq.consumer.Stop()
}

func (nsq nsq) connectToLookupd(addr string) error {
	return nsq.consumer.ConnectToNSQLookupd(addr)
}
