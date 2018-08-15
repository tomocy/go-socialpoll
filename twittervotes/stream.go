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

type stream interface {
	open()
	close()
	closeConnection()
}

type twitterStream struct {
	setUpOnce  sync.Once
	authCreds  *oauth.Credentials
	client     *http.Client
	conn       net.Conn
	readCloser io.ReadCloser
	db         db
	voteCh     chan<- string
	closeCh    chan struct{}
	closedCh   chan<- struct{}
}

func newStream(dbURL string, voteCh chan<- string, closedCh chan<- struct{}) stream {
	return newTwitterStream(dbURL, voteCh, closedCh)
}

func newTwitterStream(dbURL string, voteCh chan<- string, closedCh chan<- struct{}) *twitterStream {
	return &twitterStream{
		db:       newTwitterVoteDB(dbURL),
		voteCh:   voteCh,
		closeCh:  make(chan struct{}),
		closedCh: closedCh,
	}
}

func (s *twitterStream) open() {
	if err := s.db.dial(); err != nil {
		log.Fatalf("twitterStream could not dial DB: %s\n", err)
	}
	defer s.db.close()

	for {
		select {
		case <-s.closeCh:
			s.closedCh <- struct{}{}
			log.Println("twitterStream closed")
			return
		default:
			log.Println("twitterStream started connecting to twitter")
			s.read()
			log.Println("twitterStream is waiting for 10s for next request")
			time.Sleep(10 * time.Second)
		}
	}
}

func (s *twitterStream) read() {
	options, err := s.db.loadOptions()
	if err != nil {
		log.Printf("twitterStream could not load options from DB: %s\n", err)
		return
	}
	resp, err := s.makeRequestToDetectTweetsRegardingVote(options)
	if err != nil {
		log.Printf("twitterStream could not make request to detect tweets redarding options: %s\n", err)
		return
	}

	s.readCloser = resp.Body
	go s.deliverOptions(options)
}

func (s *twitterStream) makeRequestToDetectTweetsRegardingVote(options []string) (*http.Response, error) {
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	return s.makeRequestForStreaming(query)
}

func (s *twitterStream) makeRequestForStreaming(query url.Values) (*http.Response, error) {
	req, err := http.NewRequest(
		"POST",
		"https://stream.twitter.com/1.1/statuses/filter.json",
		strings.NewReader(query.Encode()),
	)
	if err != nil {
		log.Printf("twitterStream could not get new request instance: %s\n", err)
	}

	return s.makeRequest(req, query)
}

func (s *twitterStream) makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	s.setUpOnce.Do(func() {
		s.setUpTwitterAuth()
		s.client = &http.Client{
			Transport: &http.Transport{
				Dial: s.dial,
			},
		}
	})

	encodedParams := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(encodedParams)))

	authClient := new(oauth.Client)
	req.Header.Set("Authorization", authClient.AuthorizationHeader(s.authCreds, "POST", req.URL, params))

	return s.client.Do(req)
}

func (s *twitterStream) setUpTwitterAuth() {
	var authCreds = struct {
		ClientKey    string `env:"SP_TWITTER_CLIENT_KEY,required"`
		ClientSecret string `env:"SP_TWITTER_CLIENT_SECRET,required"`
	}{}
	if err := envdecode.Decode(&authCreds); err != nil {
		log.Fatalf("twitterStream could not decode twitter credentials from env: %s\n", err)
	}

	s.authCreds = &oauth.Credentials{
		Token:  authCreds.ClientKey,
		Secret: authCreds.ClientSecret,
	}
}

func (s *twitterStream) dial(network, addr string) (net.Conn, error) {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	newConn, err := net.DialTimeout(network, addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	s.conn = newConn
	return newConn, nil
}

func (s *twitterStream) close() {
	s.closeConnection()
	s.closeCh <- struct{}{}
}

func (s *twitterStream) closeConnection() {
	if s.conn != nil {
		s.conn.Close()
	}

	if s.readCloser != nil {
		s.readCloser.Close()
	}
}

type tweet struct {
	Text string
}

func (s twitterStream) deliverOptions(options []string) {
	decoder := json.NewDecoder(s.readCloser)
	for {
		var tweet tweet
		if err := decoder.Decode(&tweet); err != nil {
			log.Println("twitterStream finished decoding")
			break
		}

		for _, option := range options {
			log.Println(option)
			if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(option)) {
				log.Println("vote: ", option)
				s.voteCh <- option
			}
		}
	}
}
