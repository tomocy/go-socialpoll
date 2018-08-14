package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	nsq "github.com/bitly/go-nsq"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var updateDuration = 1 * time.Second

func main() {
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	dbURL := "mongodb://ballots:ballots@localhost/ballots"
	log.Println("connecting to DB: ", dbURL)
	db, err := mgo.Dial(dbURL)
	if err != nil {
		fatal(err)
		return
	}
	defer func() {
		log.Println("closing connection to db")
		db.Close()
	}()

	log.Println("connecting to NSQ")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}
	var countsLock sync.Mutex
	var counts map[string]int
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		option := string(m.Body)
		counts[option]++

		return nil
	}))

	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return
	}

	log.Println("waiting for votes on NSQ")
	polls := db.DB("ballots").C("polls")
	var updater *time.Timer
	updater = time.AfterFunc(updateDuration, func() {
		countsLock.Lock()
		defer countsLock.Unlock()

		if len(counts) == 0 {
			log.Println("skip counting votes because of no data")
		} else {
			log.Println("updating counts in db")
			log.Println(counts)
			ok := true
			for option, count := range counts {
				selector := bson.M{
					"options": bson.M{
						"$in": []string{option},
					},
				}
				update := bson.M{
					"$inc": bson.M{
						"results." + option: count,
					},
				}
				if _, err := polls.UpdateAll(selector, update); err != nil {
					log.Printf("failed to update: %s\n", err)
					ok = false
					continue
				}
			}
			if ok {
				log.Println("update counts in db successfully")
				counts = nil
			}
		}

		updater.Reset(updateDuration)
	})

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-signalCh:
			updater.Stop()
			q.Stop()
		case <-q.StopChan:
			return
		}
	}
}

var fatalErr error

func fatal(e error) {
	log.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}
