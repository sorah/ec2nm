package main

import (
	"os"
	"net"
	"fmt"
	"strings"
	"flag"

	"github.com/miekg/dns"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
)

type Handl struct {
	Instances map[string]ec2.Instance;
}

func (hn *Handl) ServeDNS(writer dns.ResponseWriter, req *dns.Msg) {
	fmt.Println(strings.Split(req.Question[0].Name, ".")[0])

	response := new(dns.Msg)
	response.SetReply(req)
	response.Authoritative = true

	instance, ok := hn.Instances[strings.Split(req.Question[0].Name, ".")[0]]
	if ! ok { // empty
		fmt.Println("Not found")
		writer.WriteMsg(response)
		return
	}

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
	rr.A = net.ParseIP(instance.PrivateIpAddress)

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
	fmt.Println("Hello!")

	regionName := flag.String("region", os.Getenv("AWS_REGION"), "AWS Region name")

	flag.Parse()

	if *regionName == "" {
		panic("Region should be specified via -region option or $AWS_REGION")
	}
	region := aws.Regions[*regionName]

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
	handler.Instances = make(map[string]ec2.Instance, 10)

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if tag.Key == "Name" {
					fmt.Printf("%v: %v\n", tag.Value, net.ParseIP(instance.PrivateIpAddress))
					handler.Instances[tag.Value] = instance
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

	fmt.Println("Start...")
	server.ListenAndServe()
}
