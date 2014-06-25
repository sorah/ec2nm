package main

import (
	"os"
	"net"
	"fmt"
	"strings"
	"flag"
	"time"

	"github.com/miekg/dns"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
)

type Handler struct {
	Domain string;
	Instances map[string]ec2.Instance;
	TimeToAlive uint;
}

func (hn *Handler) ServeDNS(writer dns.ResponseWriter, req *dns.Msg) {
	response := new(dns.Msg)
	response.SetReply(req)

	domainIndex := strings.LastIndex(req.Question[0].Name, hn.Domain)
	if domainIndex < 0 {
		fmt.Println("Not target domain")
		writer.WriteMsg(response)

		return
	}

	queryName := req.Question[0].Name[0:(domainIndex-1)]
	paths := strings.Split(queryName, ".")

	instanceName := paths[len(paths)-1]

	instance, ok := hn.Instances[instanceName]
	if ! ok { // empty
		fmt.Println("Not found")
		writer.WriteMsg(response)
		return
	}

	response.Authoritative = true

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(hn.TimeToAlive)}
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

func UpdateInstances(client *ec2.EC2, handler *Handler) {
	fmt.Println("Updating instances...")
	instances, err := client.Instances(nil, nil)
	if err != nil {
		panic(err)
	}

	newInstances := make(map[string]ec2.Instance, 10)

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if tag.Key == "Name" {
					fmt.Printf("%v: %v\n", tag.Value, net.ParseIP(instance.PrivateIpAddress))
					newInstances[tag.Value] = instance
				}
			}
		}
	}

	handler.Instances = newInstances
	fmt.Println("----")
}

func PeriodicalInstanceUpdater(interval int, client *ec2.EC2, handler *Handler) {
	for _ = range time.Tick(time.Duration(interval) * time.Second) {
		UpdateInstances(client, handler)
	}
}

func main() {
	fmt.Println("Hello!")

	regionName := flag.String("region", os.Getenv("AWS_REGION"), "AWS Region name")
	domain := flag.String("domain", "aws", "Suffix for instance records")
	ttl := flag.Uint("ttl", 280, "TTL for DNS records")
	interval := flag.Int("interval", 300, "Interval to update Instances data")
	bind := flag.String("bind", ":10053", "bind address + port")
	protocol := flag.String("protocol", "udp", "protocol")

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

	handler := new(Handler)
	handler.Domain = *domain
	handler.TimeToAlive = *ttl

	UpdateInstances(client, handler)
	go PeriodicalInstanceUpdater(*interval, client, handler)

	mux := dns.NewServeMux()
	mux.Handle(".", handler)

	server := &dns.Server{
		Addr: *bind,
		Handler: mux,
		Net: *protocol,
	}

	fmt.Println("Start...")
	server.ListenAndServe()
}
