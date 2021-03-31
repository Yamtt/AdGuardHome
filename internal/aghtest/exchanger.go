package aghtest

import (
	"net"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

// Exchanger is a mock aghnet.Exchanger implementation for tests.
type Exchanger struct {
	Ups upstream.Upstream
}

// Exchange implements aghnet.Exchanger interface for *Exchanger.
func (e *Exchanger) Exchange(req *dns.Msg) (resp *dns.Msg, err error) {
	if e.Ups == nil {
		e.Ups = &TestErrUpstream{}
	}

	return e.Ups.Exchange(req)
}

// RDNSExchanger is a mock dnsforward.RDNSExchanger implementation for tests.
type RDNSExchanger struct {
	Exchanger
}

// Exchange implements dnsforward.RDNSExchanger interface for *RDNSExchanger.
func (e *RDNSExchanger) Exchange(ip net.IP) (host string, err error) {
	req := &dns.Msg{
		Question: []dns.Question{{
			Name:  ip.String(),
			Qtype: dns.TypePTR,
		}},
	}

	resp, err := e.Exchanger.Exchange(req)
	if err != nil {
		return "", err
	}

	if len(resp.Answer) == 0 {
		return "", nil
	}

	return resp.Answer[0].Header().Name, nil
}
