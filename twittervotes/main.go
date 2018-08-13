package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	nsq "github.com/bitly/go-nsq"
	mgo "gopkg.in/mgo.v2"
)

func main() {
	var stopLock sync.Mutex
	isStop := false
	stopCh := make(chan struct{})
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		stopLock.Lock()
		isStop = true
		stopLock.Unlock()

		log.Println("stopping")
		stopCh <- struct{}{}
		closeConn()
	}()

	if err := dialDB(); err != nil {
		log.Fatalf("could not dial to DB: %s\n", err)
	}
	defer closeDB()

	votes := make(chan string)
	publisherStopCh := publishVotes(votes)
	twitterStreamStopCh := startTwitterStream(stopCh, votes)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			closeConn()
			stopLock.Lock()
			if isStop {
				stopLock.Unlock()
				return
			}
			stopLock.Unlock()
		}
	}()

	<-twitterStreamStopCh
	close(votes)
	<-publisherStopCh
}

var db *mgo.Session

func dialDB() error {
	log.Println("dial to MongoDB: localhost")
	var err error
	db, err = mgo.Dial("localhost")
	return err
}

func closeDB() {
	log.Println("connection to db is closed")
	db.Close()
}

type poll struct {
	Options []string
}

func loadOptions() ([]string, error) {
	options := make([]string, 0)
	var p poll
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}

	return options, iter.Err()
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
