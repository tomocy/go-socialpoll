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
	twitterStreamStopCh chan struct{}
	isStopLocker        sync.Mutex
	isStopped           bool
}

func newTwitterVote(twitterStreamStopCh chan struct{}) *twitterVote {
	termSignalCh := make(chan os.Signal)
	signal.Notify(termSignalCh, syscall.SIGINT)
	return &twitterVote{
		termSignalCh:        termSignalCh,
		twitterStreamStopCh: twitterStreamStopCh,
	}
}

func (v *twitterVote) waitInterruptSignalToFinishTwitterStream() {
	log.Println("twitter vote is waiting interrupt signal to finish twitter stream")
	<-v.termSignalCh
	log.Println("twitter vote is stopping and finishing twitter stream")
	v.stop()
	v.sendStopSignalToTwitterStream()
	closeConn()
}

func (v *twitterVote) stop() {
	v.isStopLocker.Lock()
	v.isStopped = true
	v.isStopLocker.Unlock()
}

func (v *twitterVote) sendStopSignalToTwitterStream() {
	v.twitterStreamStopCh <- struct{}{}
}

func (v *twitterVote) closeConnectionToTwitterStreamPerSecond() {
	log.Println("twitter vote closes connection to twitter stream per second")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			closeConn()
		}

		if v.doesStop() {
			break
		}
	}
}

func (v *twitterVote) doesStop() bool {
	v.isStopLocker.Lock()
	defer v.isStopLocker.Unlock()
	return v.isStopped
}
