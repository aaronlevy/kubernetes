/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restclient

import (
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	"github.com/golang/glog"
)

func newSlightlyStickyProvider(hosts []*url.URL) *slightlyStickyProvider {
	return &slightlyStickyProvider{
		hosts: hosts,
		cur:   rand.Intn(len(hosts)),
	}
}

type slightlyStickyProvider struct {
	sync.RWMutex
	hosts []*url.URL
	cur   int
}

func (s *slightlyStickyProvider) get() *url.URL {
	s.RLock()
	defer s.RUnlock()
	glog.Infof("XXX: provider get: %d:%s", s.cur, s.hosts[s.cur].String())
	return s.hosts[s.cur]
}

func (s *slightlyStickyProvider) next() {
	s.Lock()
	defer s.Unlock()
	s.cur = (s.cur + 1) % len(s.hosts)
}

func (s *slightlyStickyProvider) wrap(delegate http.RoundTripper) http.RoundTripper {
	// Fast-path of original single-host functionality
	if len(s.hosts) == 1 {
		return delegate
	}
	return rtfunc(func(req *http.Request) (*http.Response, error) {
		resp, err := delegate.RoundTrip(req)
		if err != nil {
			glog.Infof("XXX: delegate err: %v", err)
			s.next()
		}
		return resp, err
	})
}

type rtfunc func(*http.Request) (*http.Response, error)

func (rt rtfunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}
