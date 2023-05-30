package main

import (
	"fpco-internal/pgocomp/pkg/awsc/awscnet"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return awscnet.NewBasicNetworkComponent(
			"my-network",
			awscnet.BasicNetworkParameters{
				Vpc: awscnet.CreateVpcParameters{
					Region:    "us-east-2",
					CidrBlock: "10.0.0.0/16",
				},
				PublicSubnet: awscnet.CreateSubnetParameters{
					AvailabilityZone: "us-east-2a",
					CidrBlock:        "10.0.0.0/24",
				},
				PrivateSubnet: awscnet.CreateSubnetParameters{
					AvailabilityZone: "us-east-2b",
					CidrBlock:        "10.0.1.0/24",
				},
			}).Apply(ctx)
	})
}
