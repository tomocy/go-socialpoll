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
	termSignalCh        chan os.Signal
	stream              *twitterStream
	twitterStreamStopCh chan struct{}
	isStoppedLocker     sync.Mutex
	isStopped           bool
}

func newTwitterVote(db twitterVoteDB, twitterStreamStopCh chan struct{}) *twitterVote {
	termSignalCh := make(chan os.Signal)
	signal.Notify(termSignalCh, syscall.SIGINT)
	return &twitterVote{
		termSignalCh:        termSignalCh,
		stream:              newTwitterStream(db),
		twitterStreamStopCh: twitterStreamStopCh,
	}
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
