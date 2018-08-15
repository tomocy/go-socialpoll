package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type twitterVote struct {
	stream              *twitterStream
	nsq                 *twitterVoteNSQ
	db                  *twitterVoteDB
	termSignalCh        chan os.Signal
	stoppedCh           chan struct{}
	twitterStreamStopCh chan struct{}
	votesCh             chan string
	isStoppedLocker     sync.Mutex
	isStopped           bool
}

func newTwitterVote(dbURL string, nsqURL string) *twitterVote {
	termSignalCh := make(chan os.Signal)
	signal.Notify(termSignalCh, syscall.SIGINT)
	// db := newTwitterVoteDB(dbURL)
	return &twitterVote{
		stream:              newTwitterStream(),
		nsq:                 newTwitterVoteNSQ(nsqURL),
		db:                  newTwitterVoteDB(dbURL),
		termSignalCh:        termSignalCh,
		twitterStreamStopCh: make(chan struct{}),
		votesCh:             make(chan string),
	}
}

func (v *twitterVote) start() {
	log.Println("twitterVote started")
	go v.waitInterruptSignalToFinishTwitterStream()
	go v.closeConnectionToTwitterStreamPerSecond()

	v.dialDB()
	defer v.closeDB()

	go v.publishVotes()
	go v.startStream()

	v.waitForStreamAndNSQToBeClosed()
}

func (v *twitterVote) waitInterruptSignalToFinishTwitterStream() {
	log.Println("twitterVote is waiting interrupt signal to finish twitter stream")
	<-v.termSignalCh
	log.Println("twitterVote is stopping and finishing twitter stream")
	v.stop()
	v.stream.close()
	v.notifyOfHavingClosedStream()
}

func (v *twitterVote) stop() {
	v.isStoppedLocker.Lock()
	v.isStopped = true
	v.isStoppedLocker.Unlock()
}

func (v *twitterVote) notifyOfHavingClosedStream() {
	v.twitterStreamStopCh <- struct{}{}
}

func (v *twitterVote) closeConnectionToTwitterStreamPerSecond() {
	log.Println("twitterVote closes connection to twitter stream per second")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			v.stream.close()
		}

		if v.doesStop() {
			break
		}
	}
}

func (v *twitterVote) doesStop() bool {
	v.isStoppedLocker.Lock()
	defer v.isStoppedLocker.Unlock()
	return v.isStopped
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

func (v twitterVote) startStream() {
	options, err := v.db.loadOptions()
	if err != nil {
		log.Fatalf("could not load options from db: %s\n", err)
	}
	v.stream.start(v.twitterStreamStopCh, v.votesCh, options)
}

func (v twitterVote) waitForStreamAndNSQToBeClosed() {
	<-v.stream.stoppedCh
	close(v.votesCh)
	<-v.nsq.stoppedCh
}
