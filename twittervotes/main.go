package main

func main() {
	twitterVote := newTwitterVote("mongodb://ballots:ballots@localhost/ballots", "localhost:4150")
	twitterVote.start()
}
