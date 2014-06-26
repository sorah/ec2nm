# ec2nm - DNS server helps resolving your EC2 instances' IP addresses

## Install

```
$ go get github.com/sorah/ec2nm
```

## Run

```
$ AWS_ACCESS_KEY_ID=XXX AWS_SECRET_ACCESS_KEY=XXX ec2nm -region ap-northeast-1
```

## Usage

(By default, port is `10053` and domain (record suffix) is `aws`)

```
$ dig +short -p 10053 @localhost instance.aws
10.x.x.x (private ip for Name:instance)

$ dig +short -p 10053 @localhost instance.vpc-deadbeef.aws
10.x.x.x (private ip for Name:instance in vpc-deadbeef)

$ dig +short -p 10053 @localhost instance.ap-northeast-1.aws
10.x.x.x (private ip for Name:instance, located in ap-northeast-1)

$ dig +short -p 10053 @localhost instance.vpc-deadbeef.ap-northeast-1.aws
10.x.x.x (private ip for Name:instance in vpc-deadbeef, located in ap-northeast-1)
```

## Options


### `-region`

Specify AWS region names. It can specify multiple values (separeted by comma.)

```
-region ap-northeast-1,us-west-1
```

### `-bind`

Specify bind address and port. Default `:10053`.

### `-interval`

Interval to update instance lists in seconds. Default `300` seconds.

### `-ttl`

TTL for records. Default `280` seconds.

### `-vpc-aliases`

Set aliases for VPCs.  It can specify multiple values (separeted by comma.)

```
$ ec2nm -vpc-aliases vpc-deadbeef:my-vpc,vpc-8badf00d:another-vpc
```

By this example, `my-vpc` effects as an alias to `vpc-deadbeef`, `another-vpc` effects as an alias to `vpc-8badf00d`.

## IAM policy

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

## License

```
Copyright (c) 2014 Shota Fukumori (sora_h)

MIT License

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```
