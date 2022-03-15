package main

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

type upstream struct {
	addr string
}

func tryAddPort(addr string, port string) string {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, port)
	}
	return addr
}

func newUpstream(addr string) *upstream {
	return &upstream{addr: tryAddPort(addr, "53")}
}

func (u *upstream) exchange(m *dns.Msg) (*dns.Msg, error) {
	if m.IsEdns0() != nil {
		return u.exchangeOPTM(m)
	}
	mc := m.Copy()
	mc.SetEdns0(512, false)
	r, err := u.exchangeOPTM(mc)
	if err != nil {
		return nil, err
	}
	removeEDNS0(r)
	return r, nil
}

func (u *upstream) exchangeOPTM(m *dns.Msg) (*dns.Msg, error) {
	c, err := dns.Dial("udp", u.addr)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(time.Second * 3))
	if opt := m.IsEdns0(); opt != nil {
		c.UDPSize = opt.UDPSize()
	}
	if err := c.WriteMsg(m); err != nil {
		return nil, err
	}

	for {
		r, err := c.ReadMsg()
		if err != nil {
			return nil, err
		}
		if r.IsEdns0() == nil {
			continue
		}
		return r, nil
	}
}

func removeEDNS0(m *dns.Msg) {
	for i := len(m.Extra) - 1; i >= 0; i-- {
		if m.Extra[i].Header().Rrtype == dns.TypeOPT {
			m.Extra = append(m.Extra[:i], m.Extra[i+1:]...)
			return
		}
	}
	return
}
