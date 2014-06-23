package main

import (
	"fmt"
	"github.com/miekg/dns"
)

func Serve(writer dns.ResponseWriter, req *dns.Msg) {
	fmt.Println(req.Question[0].Name)

	response := new(dns.Msg)
	response.SetReply(req)
	response.Authoritative = true

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: "test.internal.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
	rr.A = []byte{10,1,2,3}

	switch req.Question[0].Qtype {
	default:
		fallthrough
	case dns.TypeAAAA, dns.TypeA:
		response.Answer = []dns.RR{rr}
		// response.Extra =
	}

	writer.WriteMsg(response)
}

func main() {
	fmt.Println("Hello")

	mux := dns.NewServeMux()
	mux.HandleFunc(".", Serve)

	server := &dns.Server{
		Addr: ":10053",
		Handler: mux,
		Net:  "udp",
	}

	server.ListenAndServe()
}
