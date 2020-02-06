package spider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Spider struct {
	URL                URL
	SpiderDepth        int
	Goroutines         int
	ActionTimeout      time.Duration
	AllowedDomain      Domain
	AllowedContentType ContentType

	PreAction      func(req *http.Request) bool
	PostAction     func(req *http.Request, resp *http.Response, body *bytes.Buffer) bool
	ExpandPolicy   func(req *http.Request, body *bytes.Buffer) URL
	RedirectPolicy func(req *http.Request, via []*http.Request) error
	OnError        func(req *http.Request, err error)

	tokens chan struct{}
	sync   sync.WaitGroup
	seen   map[string]bool

	mu     sync.Mutex
	expand []string
}

func (s *Spider) check() error {
	if len(s.URL) <= 0 {
		return fmt.Errorf("spider url empty")
	}
	if s.SpiderDepth < 0 {
		return fmt.Errorf("spider depth %d", s.SpiderDepth)
	}
	if s.Goroutines <= 0 {
		s.Goroutines = 1
	}
	if s.ExpandPolicy == nil {
		s.ExpandPolicy = defaultExpandPolicy
	}
	if len(s.AllowedDomain) <= 0 {
		for _, u := range s.URL {
			host, err := url.Parse(u)
			if err != nil {
				return err
			}
			s.AllowedDomain = append(s.AllowedDomain, host.Hostname())
		}
	}
	if len(s.AllowedContentType) <= 0 {
		s.AllowedContentType = append(s.AllowedContentType, HtmlType)
	}
	if s.OnError == nil {
		errlog := log.New(os.Stderr, "Spider", log.Lshortfile)
		s.OnError = func(req *http.Request, err error) {
			errlog.Println("Error", req.URL, err)
		}
	}
	if s.PreAction == nil {
		s.PreAction = func(req *http.Request) bool {
			return false
		}
	}
	if s.PostAction == nil {
		s.PostAction = func(req *http.Request, resp *http.Response, body *bytes.Buffer) bool {
			return false
		}
	}

	return nil
}

func (s *Spider) Run() error {

	if err := s.check(); err != nil {
		return err
	}

	s.seen = make(map[string]bool)
	s.tokens = make(chan struct{}, s.Goroutines)

	links := s.URL
	for ; s.SpiderDepth >= 0; s.SpiderDepth-- {
		for _, link := range links {
			s.sync.Add(1)
			go s.run(link)
		}
		s.sync.Wait()
		links = nil
		for _, l := range s.expand {
			if !s.seen[l] {
				s.seen[l] = true
				links = append(links, l)
			}
		}
	}
	return nil
}

func (s *Spider) run(url string) {
	defer s.sync.Done()
	s.tokens <- struct{}{}
	defer func() { <-s.tokens }()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.OnError(req, err)
		return
	}
	if s.PreAction(req) {
		return
	}

	resp, err := (&http.Client{
		CheckRedirect: s.RedirectPolicy,
		Timeout:       s.ActionTimeout,
	}).Do(req)
	if err != nil {
		s.OnError(req, err)
		return
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		s.OnError(req, err)
		return
	}
	_ = resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if s.checkContentType(ct) {
		return
	}

	if s.PostAction(req, resp, bytes.NewBuffer(bodyBytes)) {
		return
	}

	if strings.Contains(ct, "text/html") {
		s.mu.Lock()
		s.expand = appendSlice(s.expand, s.checkDomain, s.ExpandPolicy(req, bytes.NewBuffer(bodyBytes))...)
		s.mu.Unlock()
	}
}

