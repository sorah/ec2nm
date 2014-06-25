package main

import (
	"os"
	"fmt"
	"strings"
	"flag"
	"time"

	"github.com/miekg/dns"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"

	"github.com/sorah/ec2nm/ec2nm"
)

func PeriodicalInstanceUpdater(interval int, handler *ec2nm.Handler) {
	for _ = range time.Tick(time.Duration(interval) * time.Second) {
		handler.UpdateInstances()
	}
}

func parseVpcAliasesOption(str string) map[string]string {
	ary := strings.Split(str, ",")
	aliases := make(map[string]string, len(ary))
	for _, aliasStr := range ary {
		alias := strings.Split(aliasStr, ":")
		fmt.Println(aliasStr)
		if len(alias) == 2 {
			aliases[alias[1]] = alias[0]
		}
	}
	return aliases
}

func main() {
	fmt.Println("Hello!")


	regionName := flag.String("region", os.Getenv("AWS_REGION"), "AWS Region name")
	domain := flag.String("domain", "aws", "Suffix for instance records")
	ttl := flag.Uint("ttl", 280, "TTL for DNS records")
	interval := flag.Int("interval", 300, "Interval to update Instances data")
	bind := flag.String("bind", ":10053", "bind address + port")
	protocol := flag.String("protocol", "udp", "protocol")
	vpcAliases := flag.String("vpc-aliases", "", "Alias for VPC id (ex: vpc-deadbeef:myvpc,vpc-...)")

	flag.Parse()

	if *regionName == "" {
		panic("Region should be specified via -region option or $AWS_REGION")
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	config := &ec2nm.Config{
		EC2Clients: make(map[string]*ec2.EC2, 1),
		RegionNames: strings.Split(*regionName, ","),
		AWSCredential: auth,
		Domain: *domain,
		TTL: uint32(*ttl),
		Interval: *interval,
		Bind: *bind,
		Protocol: *protocol,
		VpcAliases: parseVpcAliasesOption(*vpcAliases),
	}

	handler := &ec2nm.Handler{
		Config: config,
		Regions: make(map[string]map[string]ec2.Instance, 1),
		VpcsInRegions: make(map[string]map[string]*map[string]ec2.Instance, 1),
	}

	handler.UpdateInstances()
	go PeriodicalInstanceUpdater(*interval, handler)

	mux := dns.NewServeMux()
	mux.Handle(".", handler)

	server := &dns.Server{
		Addr: config.Bind,
		Net: config.Protocol,
		Handler: mux,

	}

	fmt.Println("Start...")
	server.ListenAndServe()
}
