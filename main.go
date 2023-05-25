package main

import (
	flag "github.com/spf13/pflag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
)

// adapted from https://gist.github.com/NinoM4ster/edaac29339371c6dde7cdb48776d2854 which was
// adapted from https://gist.github.com/walm/0d67b4fb2d5daf3edd4fad3e13b162cb
// to support multiple A records (different IPs) and multiple SRV records (same host, different ports).

func newDNSHandler(aRecords, aaaaRecords []net.IP, aDelay, aaaaDelay time.Duration, authority string) dns.HandlerFunc {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Compress = false

		if r.Opcode != dns.OpcodeQuery {
			log.Printf("Got a non-query message: %v\n", r)
			w.WriteMsg(m)
			return
		}

		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				log.Printf("A Query for %s, replying aftter %v\n", q.Name, aDelay)
				time.Sleep(aDelay)
				if len(aRecords) > 0 {
					for _, ip := range aRecords {
						rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
						if err == nil {
							m.Answer = append(m.Answer, rr)
						}
					}
				}

			case dns.TypeAAAA:
				log.Printf("AAAA Query for %s, replying aftter %v\n", q.Name, aaaaDelay)
				time.Sleep(aaaaDelay)
				if len(aaaaRecords) > 0 {
					for _, ip := range aaaaRecords {
						rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ip))
						if err == nil {
							m.Answer = append(m.Answer, rr)
						}
					}
				}
			}
		}

		if len(authority) > 0 {
			rr, err := dns.NewRR(fmt.Sprintf("%s NS %s", m.Question[0].Name, authority))
			if err == nil {
				m.Ns = append(m.Ns, rr)
			}
		}
		m.Authoritative = true

		w.WriteMsg(m)
	}
}

func main() {
	// Flags
	port := flag.IntP("port", "p", 5353, "port to listen on")
	listenAddr := flag.StringP("listen", "l", "0.0.0.0", "address to listen on")
	aRecords := flag.IPSliceP("a", "a", []net.IP{}, "A records to serve")
	aaaaRecords := flag.IPSliceP("aaaa", "6", []net.IP{}, "AAAA records to serve")
	aDelay := flag.DurationP("delay-a", "d", 0, "delay before serving to A records")
	aaaaDelay := flag.DurationP("delay-aaaa", "D", 0, "delay before serving to AAAA records")
	authority := flag.StringP("authority", "", "", "authority to serve")
	flag.Parse()

	// attach request handler func
	dns.HandleFunc(".", newDNSHandler(*aRecords, *aaaaRecords, *aDelay, *aaaaDelay, *authority))

	// start server
	server := &dns.Server{Addr: fmt.Sprintf("%s:%d", *listenAddr, *port), Net: "udp"}
	log.Printf("Starting at %s\n", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

	defer server.Shutdown()
}
