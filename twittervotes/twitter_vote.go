package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type twitterVote struct {
	termSignalCh        chan os.Signal
	twitterStreamStopCh chan struct{}
	isStopLocker        sync.Mutex
	isStop              bool
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
	v.isStop = true
	v.isStopLocker.Unlock()
}

func (v *twitterVote) sendStopSignalToTwitterStream() {
	v.twitterStreamStopCh <- struct{}{}
}
