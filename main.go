package main

import (
	"os"
	"net"
	"fmt"
	"strings"
	"github.com/miekg/dns"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
)

type Handl struct {
	Instances map[string]net.IP;
}

func (hn *Handl) ServeDNS(writer dns.ResponseWriter, req *dns.Msg) {
	fmt.Println(strings.Split(req.Question[0].Name, ".")[0])

	response := new(dns.Msg)
	response.SetReply(req)
	response.Authoritative = true

	instance := hn.Instances[strings.Split(req.Question[0].Name, ".")[0]]
	if instance == nil {
		writer.WriteMsg(response)
		return
	}

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
	rr.A = instance

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

	var regionName string
	if regionName = os.Getenv("AWS_REGION"); regionName == "" {
	  regionName = "ap-northeast-1"
	}
	region := aws.Regions[regionName]

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	client := ec2.New(auth, region)
	instances, err := client.Instances(nil, nil)
	if err != nil {
		panic(err)
	}

	handler := new(Handl)
	handler.Instances = make(map[string]net.IP, 10)

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if tag.Key == "Name" {
					fmt.Printf("%v -- %v\n", tag.Value, net.ParseIP(instance.PrivateIpAddress))
					handler.Instances[tag.Value] = net.ParseIP(instance.PrivateIpAddress)
				}
			}
		}
	}

	mux := dns.NewServeMux()
	mux.Handle(".", handler)

	server := &dns.Server{
		Addr: ":10053",
		Handler: mux,
		Net:  "udp",
	}

	server.ListenAndServe()
}
