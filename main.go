package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/miekg/dns"
)

// adapted from https://gist.github.com/NinoM4ster/edaac29339371c6dde7cdb48776d2854 which was
// adapted from https://gist.github.com/walm/0d67b4fb2d5daf3edd4fad3e13b162cb

func newDNSHandler(records Records, aDelay, aaaaDelay time.Duration, authority string) dns.HandlerFunc {
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
			queryType := ""
			answers := []string{}
			cname := false
			delay := time.Duration(0)

			if strings.HasPrefix(q.Name, "cname.") {
				cname = true
				d := strings.TrimPrefix(q.Name, "cname.")
				log.Printf("Query for %s, replying with CNAME %s\n", q.Name, d)
				rr, err := dns.NewRR(fmt.Sprintf("%s CNAME %s", q.Name, d))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
			switch q.Qtype {
			case dns.TypeA:
				if cname && len(records.CNAMEA) > 0 {
					answers = records.CNAMEA
				} else {
					answers = records.A
				}
				queryType = "A"
				delay = aDelay

			case dns.TypeAAAA:
				if cname && len(records.CNAMEAAAA) > 0 {
					answers = records.CNAMEAAAA
				} else {
					answers = records.AAAA
				}
				queryType = "AAAA"
				delay = aaaaDelay
			}

			log.Printf("%s Query for %s, replying after %v\n", queryType, q.Name, delay)
			time.Sleep(delay)
			if len(answers) > 0 {
				d := q.Name
				if cname {
					d = strings.TrimPrefix(d, "cname.")
				}
				for _, ip := range answers {
					rr, err := dns.NewRR(fmt.Sprintf("%s %s %s", d, queryType, ip))
					if err != nil {
						log.Printf("Failed to create RR: %s\n", err.Error())
						continue
					}
					m.Answer = append(m.Answer, rr)
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

type Records struct {
	A         []string
	AAAA      []string
	CNAMEA    []string
	CNAMEAAAA []string
}

func main() {
	// Flags
	port := flag.IntP("port", "p", 5353, "port to listen on")
	listenAddr := flag.StringP("listen", "l", "0.0.0.0", "address to listen on")
	aRecords := flag.StringSliceP("a", "a", []string{}, "A records to serve")
	aaaaRecords := flag.StringSliceP("aaaa", "6", []string{}, "AAAA records to serve")
	aDelay := flag.DurationP("delay-a", "d", 0, "delay before serving to A records")
	aaaaDelay := flag.DurationP("delay-aaaa", "D", 0, "delay before serving to AAAA records")
	aCname := flag.StringSliceP("cname-a", "c", []string{}, "A record to serve for CNAME queries")
	aaaaCname := flag.StringSliceP("cname-aaaa", "C", []string{}, "AAAA record to serve for CNAME queries")
	authority := flag.StringP("authority", "", "", "authority to serve")
	flag.Parse()

	records := Records{
		A:         *aRecords,
		AAAA:      *aaaaRecords,
		CNAMEA:    *aCname,
		CNAMEAAAA: *aaaaCname,
	}

	// attach request handler func
	dns.HandleFunc(".", newDNSHandler(records, *aDelay, *aaaaDelay, *authority))

	// start server
	server := &dns.Server{Addr: fmt.Sprintf("%s:%d", *listenAddr, *port), Net: "udp"}
	log.Printf("Starting at %s\n", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

	defer server.Shutdown()
}
