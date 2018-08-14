package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/joeshaw/envdecode"
)

var conn net.Conn

func dial(network, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}

	newConn, err := net.DialTimeout(network, addr, 30*time.Second)
	if err != nil {
		return nil, err
	}
	conn = newConn
	return newConn, nil
}

var reader io.ReadCloser

func closeConn() {
	if conn != nil {
		conn.Close()
	}

	if reader != nil {
		reader.Close()
	}
}

var (
	authClient *oauth.Client
	creds      *oauth.Credentials
)

func setUpTwitterAuth() {
	var twitterCreds struct {
		ClientKey    string `env:"SP_TWITTER_CLIENT_KEY,required"`
		ClientSecret string `env:"SP_TWITTER_CLIENT_SECRET,required"`
	}
	if err := envdecode.Decode(&twitterCreds); err != nil {
		log.Fatalf("could not decode twitter credentials from env: %s\n", err)
	}

	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  twitterCreds.ClientKey,
			Secret: twitterCreds.ClientSecret,
		},
	}
}

var (
	authSetUpOnce     sync.Once
	twitterAuthClient *http.Client
)

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetUpOnce.Do(func() {
		setUpTwitterAuth()
		twitterAuthClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})

	encodedParams := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(encodedParams)))
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))

	return twitterAuthClient.Do(req)
}

type tweet struct {
	Text string
}

func readFromTwitter(votes chan<- string, twitterVoteDB twitterVoteDB) {
	options, err := twitterVoteDB.loadOptions()
	if err != nil {
		log.Printf("could not load options: %s\n", err)
		return
	}
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))

	req, err := http.NewRequest(
		"POST",
		"https://stream.twitter.com/1.1/statuses/filter.json",
		strings.NewReader(query.Encode()),
	)
	if err != nil {
		log.Printf("could not make new request instance: %s\n", err)
		return
	}

	resp, err := makeRequest(req, query)
	if err != nil {
		log.Printf("could not do request: %s\n", err)
		return
	}

	reader = resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var tweet tweet
		if err := decoder.Decode(&tweet); err != nil {
			break
		}

		for _, option := range options {
			if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(option)) {
				log.Println("vote: ", option)
				votes <- option
			}
		}
	}
}

type twitterStream struct {
	db        twitterVoteDB
	stoppedCh chan struct{}
}

func initTwitterStream(db twitterVoteDB) twitterStream {
	return twitterStream{
		db:        db,
		stoppedCh: make(chan struct{}),
	}
}

func (s twitterStream) start(stopCh <-chan struct{}, votesCh chan<- string) {
	for {
		select {
		case <-stopCh:
			log.Println("twitterStream stopped streaming")
			s.stoppedCh <- struct{}{}
			return
		default:
			log.Println("twitterStream started connecting to Twitter")
			readFromTwitter(votesCh, s.db)
			log.Println("twitterStream is waiting for 10s for next request")
			time.Sleep(10 * time.Second)
		}
	}
}
