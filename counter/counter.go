package main

import (
	"log"
	"time"
)

type counter struct {
	db          db
	queue       queue
	increments  map[string]int
	incrementCh chan string
}

func newCounter(dbURL string) (*counter, error) {
	counter := &counter{
		db:          newDB(dbURL),
		increments:  make(map[string]int),
		incrementCh: make(chan string),
	}
	var err error
	counter.queue, err = newQueue(counter.incrementCh)
	if err != nil {
		return nil, err
	}

	return counter, nil
}

func (c *counter) start() error {
	if err := c.db.dial(); err != nil {
		return err
	}
	defer c.db.close()

	if err := c.queue.connectToLookupd("localhost:4161"); err != nil {
		return err
	}
	defer c.queue.stop()

	c.countAndUpdateDB()
	return nil
}

func (c *counter) countAndUpdateDB() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if len(c.increments) == 0 {
				log.Println("counter skipped update of the count because of no counts data")
				break
			}
			c.updateCountsInDB()
		case option := <-c.incrementCh:
			c.increaseCount(option)
		}
	}
}

func (c *counter) updateCountsInDB() {
	c.db.updateCounts(c.increments)
	c.increments = nil
}

func (c *counter) increaseCount(option string) {
	if c.increments == nil {
		c.increments = make(map[string]int)
	}
	c.increments[option]++
}
