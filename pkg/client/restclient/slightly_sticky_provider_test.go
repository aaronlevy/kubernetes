package restclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

type testRoundTripper struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.Request = req
	return rt.Response, rt.Err
}

func TestSlightyStickyWrappedRoundTrip(t *testing.T) {
	trt := &testRoundTripper{}
	errRT := &testRoundTripper{Err: errors.New("expected")}

	for i, tc := range []struct {
		hosts    int
		index    int
		expected int
		rt       http.RoundTripper
	}{
		{hosts: 3, index: 0, expected: 0, rt: trt}, // No errors - no host change
		{hosts: 3, index: 1, expected: 1, rt: trt},
		{hosts: 3, index: 0, expected: 1, rt: errRT}, // error in round-trip - next host
		{hosts: 3, index: 1, expected: 2, rt: errRT},
		{hosts: 3, index: 2, expected: 0, rt: errRT},
		{hosts: 1, index: 0, expected: 0, rt: errRT},
	} {
		hosts := newHosts(tc.hosts)
		p := newSlightlyStickyProvider(hosts)
		p.cur = tc.index
		p.wrap(tc.rt).RoundTrip(&http.Request{})

		if host := p.get(); host.String() != hosts[tc.expected].String() {
			t.Errorf("%d: unexpected host. Wanted=%s got=%s", i, hosts[tc.expected].String(), host.String())
		}
	}

}

func newHosts(c int) (hosts []*url.URL) {
	for i := 0; i < c; i++ {
		hosts = append(hosts, &url.URL{
			Scheme: "https",
			Host:   fmt.Sprintf("example%d", i),
		})
	}
	return hosts
}
