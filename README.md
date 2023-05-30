# pulumi-components-go

This repository hosts `pulumi-components-go`, a library that enables better organization and utilization of Pulumi Infrastructure as Code (IaC) patterns. Pulumi is a powerful tool allowing developers to define infrastructure as code using programming languages of their choice.

## Problem Statement

During an analysis of Pulumi code written in Go, it was observed that components were created alongside your functions and subsequently utilized in later component configurations. This made it difficult, especially for newcomers, to identify the dependencies each component had on others. 

Moreover, the need for a library providing basic setups was felt. While `awsx` exists for this purpose, the belief was that creating basic libraries using the 'Component' approach could lead to more concise code and easier Infrastructure as Code reviews.

## Solution

To address these challenges, the `pgocomp` library was developed. `pgocomp` turns Pulumi objects into Components, offering an easy and organized way to define infrastructure as code in Go. 

Here is a sample code snippet illustrating how you can use `pgocomp`:

```go
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
				Vpc: awscnet.VpcParameters{
					Region:    "us-east-2",
					CidrBlock: "10.0.0.0/16",
				},
				PublicSubnet: awscnet.SubnetParameters{
					AvailabilityZone: "us-east-2a",
					CidrBlock:        "10.0.0.0/24",
				},
				PrivateSubnet: awscnet.SubnetParameters{
					AvailabilityZone: "us-east-2b",
					CidrBlock:        "10.0.1.0/24",
				},
			}).Apply(ctx)
	})

}
