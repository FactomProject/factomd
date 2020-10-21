package p2p

import (
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type seed struct {
	url       string
	cache     []Endpoint
	cacheTTL  time.Duration
	cacheTime time.Time

	logger *log.Entry
}

func newSeed(url string, cacheTTL time.Duration) *seed {
	s := new(seed)
	s.url = url
	s.logger = packageLogger.WithFields(log.Fields{"subpackage": "Seed", "url": url})
	s.cacheTTL = cacheTTL
	return s
}

func (s *seed) retrieve() []Endpoint {
	if s.cache != nil && time.Since(s.cacheTime) <= s.cacheTTL {
		return s.cache
	}

	eps := make([]Endpoint, 0)
	err := WebScanner(s.url, func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		host, port, err := net.SplitHostPort(line)
		if err != nil {
			s.logger.Errorf("Badly formatted line [%s]", line)
			return
		}
		if ep, err := NewEndpoint(host, port); err != nil {
			s.logger.WithError(err).Errorf("Bad peer [%s]", line)
		} else {
			eps = append(eps, ep)
		}
	})

	if err != nil {
		s.logger.WithError(err).Errorf("unable to retrieve data from seed")
	}

	s.cacheTime = time.Now()
	s.cache = eps
	return eps
}

func (s *seed) size() int {
	s.retrieve()
	return len(s.cache)
}
