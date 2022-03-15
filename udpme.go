package main

import (
	"flag"
	"log"
	"net"

	"github.com/miekg/dns"
)

var (
	listenAddr   = flag.String("l", "127.0.0.1:5353", "listen address")
	upstreamAddr = flag.String("u", "8.8.8.8", "upstream address")
	testEDNS0    = flag.Bool("t", false, "test upstream edns0 support")
)

func main() {
	flag.Parse()

	if *testEDNS0 {
		test()
		return
	}

	run()
}

func test() {
	m := new(dns.Msg)
	m.SetQuestion("cloudflare.com.", dns.TypeA)
	m.SetEdns0(512, false)
	r, err := dns.Exchange(m, tryAddPort(*upstreamAddr, "53"))
	if err != nil {
		log.Printf("err: %s", err)
	} else {
		if opt := r.IsEdns0(); opt != nil {
			log.Printf("edns0 response received")
		} else {
			log.Printf("response received without edns0")
		}
	}
}

func run() {
	p, err := net.ListenPacket("udp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server started at %s", *listenAddr)

	u := newUpstream(*upstreamAddr)
	readBuf := make([]byte, 64*1024)
	for {
		n, addr, err := p.ReadFrom(readBuf)
		if err != nil {
			log.Fatalf("listener err: %s", err)
		}

		q := new(dns.Msg)
		if err := q.Unpack(readBuf[:n]); err != nil {
			log.Printf("invalid incoming package from [%s], err: %s", addr, err)
			continue
		}

		go func() {
			r, err := u.exchange(q)
			if err != nil {
				log.Printf("err: %s", err)
				r = new(dns.Msg)
				r.SetReply(q)
				r.Rcode = dns.RcodeServerFailure
			}
			if r != nil {
				writeBuf, err := r.Pack()
				if err != nil {
					log.Printf("failed to pack response, %s", err)
					return
				}
				p.WriteTo(writeBuf, addr)
			}
		}()
	}
}
