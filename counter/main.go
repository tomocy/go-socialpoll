package main

import "log"

func main() {
	counter, err := newCounter("mongodb://ballots:ballots@localhost/ballots")
	if err != nil {
		log.Fatalf("could not get new counter: %s\n", err)
	}
	if err := counter.start(); err != nil {
		log.Fatalf("could not start counter: %s\n", err)
	}
}
