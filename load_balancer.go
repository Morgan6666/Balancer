package Balancer

import (
	"log"
	"net/http"
	"net/http/httputil"
)

const (
	RoundRobin = "round_robin"
)

type Pool interface {
	Dispatch() node
	Complete(res *http.Response)
}

func NewPool(strategy string, hosts []string) Pool {
	switch strategy {
	case RoundRobin:
		return newRoundRobin(hosts)
	default:
		panic(Errorf("%v is not a valid strategy", strategy))
	}
}

type Balancer struct {
	*httputil.ReverseProxy
	pool Pool
}

// NewBalancer creates a new balancer to balance requests between hosts
// and uses specified strategy.
func NewBalancer(strategy string, hosts []string) *Balancer {
	b := &Balancer{
		pool: NewPool(strategy, hosts),
	}

	b.ReverseProxy = &httputil.ReverseProxy{
		Director:       b.Director,
		ModifyResponse: b.ModifyResponse,
	}

	return b
}

// Director directs the request to the node that was dispatched by pool.
func (b *Balancer) Director(r *http.Request) {
	node := b.pool.Dispatch()
	log.Println(b.pool)

	r.URL.Scheme = "http"
	r.URL.Host = node.host
}

// ModifyResponse tells the pool that the request was handled.
func (b *Balancer) ModifyResponse(res *http.Response) error {
	b.pool.Complete(res)
	return nil
}
