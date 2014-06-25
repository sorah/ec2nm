package ec2nm

import (
	"strings"
	"net"
	"fmt"

	"github.com/miekg/dns"
	"github.com/mitchellh/goamz/ec2"
)

type Handler struct {
	Config *Config;
	Instances map[string]ec2.Instance;
	Regions map[string]map[string]ec2.Instance;
	Vpcs map[string]map[string]ec2.Instance;
	VpcsInRegions map[string]map[string]*map[string]ec2.Instance;
}

// Name[-1]
// Name[-2].Region[-1]
// Name[-2].Vpc[-1]
// Name[-3].Vpc[-2].Region[-1]
func (handler *Handler) resolveInstance(path []string) (instance ec2.Instance, ok bool) {
	// Name[-1]
	instance, ok = handler.Instances[path[len(path)-1]]
	if ok {
		return instance, true
	}

	// Name[-2].Region[-1]
	instancesPerRegion, ok := handler.Regions[path[len(path)-1]]
	if ok {
		instance, ok := instancesPerRegion[path[len(path)-2]]
		if ok {
			return instance, true
		}
	}

	// Name[-2].Vpc[-1]
	instancesPerVpc, ok := handler.Vpcs[path[len(path)-1]]
	if ok {
		instance, ok := instancesPerVpc[path[len(path)-2]]
		if ok {
			return instance, true
		}
	}

	// Name[-3].Vpc[-2].Region[-1]
	vpcsPerRegion, ok := handler.VpcsInRegions[path[len(path)-1]]
	if ok {
		instancesPerVpcPtr, ok := vpcsPerRegion[path[len(path)-2]]
		if ok {
			instancesPerVpc = *instancesPerVpcPtr
			instance, ok := instancesPerVpc[path[len(path)-3]]
			if ok {
				return instance, true
			}
		}
	}

	return instance, false
}

func (handler *Handler) ServeDNS(writer dns.ResponseWriter, req *dns.Msg) {
	response := new(dns.Msg)
	response.SetReply(req)

	domainIndex := strings.LastIndex(req.Question[0].Name, handler.Config.Domain)
	if domainIndex < 0 {
		fmt.Println("Not target domain")
		writer.WriteMsg(response)

		return
	}

	queryName := req.Question[0].Name[0:(domainIndex-1)]
	path := strings.Split(queryName, ".")

	instance, found := handler.resolveInstance(path)
	if ! found {
		fmt.Println("Not found")
		writer.WriteMsg(response)
		return
	}

	response.Authoritative = true

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: handler.Config.TTL}
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

func (handler *Handler) UpdateInstances() {
	newInstances := make(map[string]ec2.Instance, 10)
	newVpcs := make(map[string]map[string]ec2.Instance, 10)
	for _, region := range handler.Config.RegionNames {
		handler.UpdateInstancesInRegion(region, &newInstances, &newVpcs)
	}
	handler.Instances = newInstances
	handler.Vpcs = newVpcs
}

func (handler *Handler) UpdateInstancesInRegion(regionName string, newInstances *map[string]ec2.Instance, newVpcs *map[string]map[string]ec2.Instance) {
	fmt.Printf("Updating instances in %s ...\n", regionName)
	client := handler.Config.EC2(regionName)

	instances, err := client.Instances(nil, nil)
	if err != nil {
		panic(err)
	}

	newInstancesPerRegion := make(map[string]ec2.Instance, 10)
	newVpcsPerRegion := make(map[string]*map[string]ec2.Instance, 1)

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if tag.Key == "Name" {
					fmt.Printf("%v: %v\n", tag.Value, net.ParseIP(instance.PrivateIpAddress))
					newInstancesPerRegion[tag.Value] = instance
					(*newInstances)[tag.Value] = instance

					if instance.VpcId != "" {
						instancesPerVpc, ok := (*newVpcs)[instance.VpcId]
						if ! ok {
							instancesPerVpc = make(map[string]ec2.Instance, 5)
							(*newVpcs)[instance.VpcId] = instancesPerVpc
							newVpcsPerRegion[instance.VpcId] = &instancesPerVpc
						}
						instancesPerVpc[tag.Value] = instance
					}
				}
			}
		}
	}

	handler.VpcsInRegions[regionName] = newVpcsPerRegion
	handler.Regions[regionName] = newInstancesPerRegion
	fmt.Println("----")
}
