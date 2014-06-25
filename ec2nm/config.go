package ec2nm

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
)

type Config struct {
	RegionNames []string;
	AWSCredential aws.Auth;
	EC2Clients map[string]*ec2.EC2;
	Domain string;
	TTL uint32;
	Interval int;
	Bind string;
	Protocol string;
	VpcAliases map[string]string;
}

func (conf *Config) EC2(region string) *ec2.EC2 {
	if client, ok := conf.EC2Clients[region]; ok {
		return client
	} else {
		client := ec2.New(conf.AWSCredential, aws.Regions[region])
		conf.EC2Clients[region] = client

		return client
	}
}

func (conf *Config)solveVpcId(str string) string {
	if vpcId, ok := conf.VpcAliases[str]; ok {
		return vpcId
	} else {
		return str
	}
}
