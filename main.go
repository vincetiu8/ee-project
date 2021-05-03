package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v3/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi/config"
)

const size = "t3.micro"

type provider struct {
	provider     *aws.Provider
	invokeOption pulumi.ResourceOrInvokeOption
}

type instanceArgs struct {
	userData       string
	amiId, groupId pulumi.StringInput
	keyName        pulumi.StringPtrInput
}

var regions = []string{
	"us-east-1",
	"ap-northeast-1",
	"eu-central-1",
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		testerZoneId, err := strconv.Atoi(conf.Require("testerZone"))
		if err != nil {
			return err
		}
		serverZoneId, err := strconv.Atoi(conf.Require("serverZone"))
		if err != nil {
			return err
		}

		testerRegion := regions[testerZoneId]
		serverRegion := regions[serverZoneId]

		opt := "available"
		var serverIp pulumi.StringOutput

		serverProvider := provider{}
		serverProvider.provider, err = aws.NewProvider(ctx, fmt.Sprintf("provider-%s", serverRegion), &aws.ProviderArgs{Region: pulumi.StringPtr(serverRegion)})
		serverProvider.invokeOption = pulumi.Provider(serverProvider.provider)

		zones, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{State: &opt}, serverProvider.invokeOption)
		if err != nil {
			return err
		}

		serverZone := zones.Names[0]

		serverArgs, err := getRegionArgs(ctx, serverRegion, serverProvider)
		if err != nil {
			return err
		}

		data, err := os.ReadFile("serve_files.sh")
		if err != nil {
			return err
		}
		serverArgs.userData = string(data)

		instance, err := deployServer(ctx, fmt.Sprintf("%v-server-node", serverZone), serverProvider, serverZone, serverArgs)
		if err != nil {
			return err
		}
		ctx.Export(fmt.Sprintf("zone %v server node ip", serverZone), instance.PublicIp)
		serverIp = instance.PublicIp

		testerArgs := serverArgs
		testerProvider := serverProvider
		testerZone := serverZone
		if testerZoneId != serverZoneId {
			testerProvider.provider, err = aws.NewProvider(ctx, fmt.Sprintf("provider-%s", testerRegion), &aws.ProviderArgs{Region: pulumi.StringPtr(testerRegion)})
			if err != nil {
				return err
			}
			testerProvider.invokeOption = pulumi.Provider(testerProvider.provider)

			zones, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{State: &opt}, testerProvider.invokeOption)
			if err != nil {
				return err
			}

			testerZone = zones.Names[0]

			testerArgs, err = getRegionArgs(ctx, testerRegion, testerProvider)
			if err != nil {
				return err
			}
		}

		data, err = os.ReadFile("test_protocols.sh")
		if err != nil {
			return err
		}
		fileName := fmt.Sprintf("server-%v-tester-%v", serverZone, testerZone)
		serverIp.ApplyT(func(ip string) (string, error) {
			testerArgs.userData = fmt.Sprintf(`
sudo apt-get -y install awscli
echo '%v' >> test_protocols.sh
chmod 777 test_protocols.sh
./test_protocols.sh %v 1 %v
./test_protocols.sh %v 10 %v`, string(data), ip, fileName, ip, fileName)

			instance, err := deployServer(ctx, fmt.Sprintf("tester-node-%v", ip), testerProvider, testerZone, testerArgs)
			if err != nil {
				return "", err
			}

			ctx.Export(fmt.Sprintf("zone %v tester %v node ip", testerZone, ip), instance.PublicIp)
			return ip, nil
		})

		return nil
	})
}

func getRegionArgs(ctx *pulumi.Context, region string, provider provider) (instanceArgs, error) {
	mostRecent := true
	ami, err := aws.GetAmi(ctx, &aws.GetAmiArgs{
		Owners: []string{"099720109477"},
		Filters: []aws.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-????????"},
			},
			{
				Name:   "state",
				Values: []string{"available"},
			},
		},
		MostRecent: &mostRecent,
	}, provider.invokeOption)
	if err != nil {
		return instanceArgs{}, err
	}

	group, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("ipfs-secgrp-%s", region), &ec2.SecurityGroupArgs{
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(22),
				ToPort:   pulumi.Int(22),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(8000),
				ToPort:   pulumi.Int(8000),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(4001),
				ToPort:   pulumi.Int(4001),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(5001),
				ToPort:   pulumi.Int(5001),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(8080),
				ToPort:   pulumi.Int(8080),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				Protocol: pulumi.String("-1"),
			},
		},
	}, provider.invokeOption)

	if err != nil {
		return instanceArgs{}, err
	}

	pubKeyData, err := ioutil.ReadFile("public.pem")

	keyPair, err := ec2.NewKeyPair(ctx, fmt.Sprintf("deployer-%s", region), &ec2.KeyPairArgs{
		PublicKey: pulumi.String(pubKeyData),
	}, provider.invokeOption)

	if err != nil {
		return instanceArgs{}, err
	}

	return instanceArgs{
		amiId:   pulumi.String(ami.Id),
		groupId: group.ID(),
		keyName: keyPair.KeyName,
	}, nil
}

func deployServer(ctx *pulumi.Context, name string, provider provider, zone string, args instanceArgs) (*ec2.Instance, error) {
	data, err := os.ReadFile("setup.sh")
	if err != nil {
		return nil, err
	}
	userData := pulumi.String(fmt.Sprintf("#!/bin/bash\n%v\n%v", string(data), args.userData))

	return ec2.NewInstance(ctx, name, &ec2.InstanceArgs{
		IamInstanceProfile: pulumi.String("S3FullAccess"),
		InstanceType:       pulumi.String(size),
		VpcSecurityGroupIds: pulumi.StringArray{
			args.groupId,
		},
		Ami:              args.amiId,
		AvailabilityZone: pulumi.String(zone),
		UserData:         userData,
		KeyName:          args.keyName,
	}, provider.invokeOption)
}
