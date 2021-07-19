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
	amiId, groupId pulumi.StringInput
	keyName        pulumi.StringPtrInput
	userData       pulumi.StringPtrOutput
}

var regions = []string{
	"us-east-1",
	"ap-east-1",
	"eu-central-1",
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		testerRegionId, err := strconv.Atoi(conf.Require("testerRegion"))
		if err != nil {
			return err
		}
		hostRegionId, err := strconv.Atoi(conf.Require("hostRegion"))
		if err != nil {
			return err
		}

		regionServers := make([]int, 3)

		regionServers[0], err = strconv.Atoi(conf.Require("region0Servers"))
		if err != nil {
			return err
		}
		regionServers[1], err = strconv.Atoi(conf.Require("region1Servers"))
		if err != nil {
			return err
		}
		regionServers[2], err = strconv.Atoi(conf.Require("region2Servers"))
		if err != nil {
			return err
		}

		var testerArgs instanceArgs
		var testerProvider provider
		var testerZone string

		opt := "available"
		createdServer := false
		var serverIp pulumi.StringOutput
		regionIds := []int{hostRegionId}

		for i := 0; i < 3; i++ {
			if i != hostRegionId {
				regionIds = append(regionIds, i)
			}
		}

		for _, i := range regionIds {
			region := regions[i]
			if regionServers[i] == 0 && testerRegionId != i {
				continue
			}

			serverProvider := provider{}
			serverProvider.provider, err = aws.NewProvider(ctx, fmt.Sprintf("provider-%s", region), &aws.ProviderArgs{Region: pulumi.StringPtr(region)})
			serverProvider.invokeOption = pulumi.Provider(serverProvider.provider)

			zones, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{State: &opt}, serverProvider.invokeOption)
			if err != nil {
				return err
			}

			serverZone := zones.Names[0]

			serverArgs, err := getRegionArgs(ctx, region, serverProvider)
			if err != nil {
				return err
			}

			for j := 0; j < regionServers[i]; j++ {
				serverName := fmt.Sprintf("server-%v-%v", region, j)
				if createdServer || i != hostRegionId {
					data, err := os.ReadFile("pin.sh")
					if err != nil {
						return err
					}

					serverArgs.userData = serverIp.ApplyT(func(ip string) string {
						return fmt.Sprintf(`
echo '%v' >> pin.sh
chmod 777 pin.sh
./pin.sh %v`, string(data), ip)
					}).(pulumi.StringOutput).ToStringPtrOutput()

					_, err = deployServer(ctx, serverName, serverProvider, serverZone, serverArgs)
					if err != nil {
						return err
					}
				} else {
					data, err := os.ReadFile("serve_files.sh")
					if err != nil {
						return err
					}
					serverArgs.userData = pulumi.StringPtr(string(data)).ToStringPtrOutput()

					instance, err := deployServer(ctx, serverName, serverProvider, serverZone, serverArgs)
					if err != nil {
						return err
					}
					serverIp = instance.PublicIp
					createdServer = true
				}
			}

			if testerRegionId == i {
				testerArgs = serverArgs
				testerProvider = serverProvider
				testerZone = serverZone
			}
		}

		data, err := os.ReadFile("test_protocols.sh")
		if err != nil {
			return err
		}
		fileName := fmt.Sprintf("test-%v-%v-%v-from-%v-to-%v", regionServers[0], regionServers[1], regionServers[2], regions[hostRegionId], regions[testerRegionId])
		numServers := regionServers[0] + regionServers[1] + regionServers[2]

		testerArgs.userData = serverIp.ApplyT(func(ip string) string {
			return fmt.Sprintf(`
sudo apt-get -y install awscli
echo '%v' >> test_protocols.sh
chmod 777 test_protocols.sh
./test_protocols.sh %v 10 %v %v`, string(data), ip, fileName, numServers)
		}).(pulumi.StringOutput).ToStringPtrOutput()

		_, err = deployServer(ctx, fmt.Sprintf("tester-%v", regions[testerRegionId]), testerProvider, testerZone, testerArgs)

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

	userData := args.userData.ApplyT(func(userData *string) string {
		return fmt.Sprintf("#!/bin/bash\n%v\n%v", string(data), *userData)
	}).(pulumi.StringOutput)

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
