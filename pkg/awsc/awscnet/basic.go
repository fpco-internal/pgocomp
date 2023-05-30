package awscnet

import (
	"fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// BasicNetworkParameters are the parameters of the function NewBasicNetworkComponent
type BasicNetworkParameters struct {
	Vpc           CreateVpcParameters
	PublicSubnet  CreateSubnetParameters
	PrivateSubnet CreateSubnetParameters
}

// BasicNetwork is the return type of the function NewBasicNetworkComponent
type BasicNetwork struct {
	Provider             *aws.Provider
	Vpc                  *ec2.Vpc
	PrivateSubnet        *ec2.Subnet
	PublicSubnet         *ec2.Subnet
	IGW                  *ec2.InternetGateway
	VpcGatewayAttachment *ec2.InternetGatewayAttachment
	RouteTable           *ec2.RouteTable
	DefaultRoute         *ec2.Route
	LBSecurityGroup      *ec2.SecurityGroup
}

// NewBasicNetworkComponent takes a name and parameters and returns a BasicNewotk Component
func NewBasicNetworkComponent(name string, params BasicNetworkParameters) *pgocomp.Component[*BasicNetwork] {
	return pgocomp.NewComponent(name,
		func(ctx *pulumi.Context) (*BasicNetwork, error) {
			var bnc BasicNetwork
			if err := CreateProvider(name+"-provider", params.Vpc.Region).GetAndThen(ctx, func(provider *aws.Provider) error {
				bnc.Provider = provider
				return CreateVPC(name+"-vpc", params.Vpc, provider).GetAndThen(ctx, func(vpc *ec2.Vpc) error {
					bnc.Vpc = vpc
					if err := CreateSubnet(name+"-public-subnet", params.PublicSubnet, vpc).GetAndThen(ctx, func(subnet *ec2.Subnet) error {
						bnc.PublicSubnet = subnet
						return nil
					}); err != nil {
						return err
					}
					if err := CreateSubnet(name+"-private-subnet", params.PrivateSubnet, vpc).GetAndThen(ctx, func(subnet *ec2.Subnet) error {
						bnc.PrivateSubnet = subnet
						return nil
					}); err != nil {
						return err
					}
					if err := CreateInternetGateway(name+"-igw", vpc).GetAndThen(ctx, func(igw *ec2.InternetGateway) error {
						bnc.IGW = igw
						return AttachInternetGatewayToVPC(name+"-igw-vpc-attachment", vpc, igw).GetAndThen(ctx, func(iga *ec2.InternetGatewayAttachment) error {
							bnc.VpcGatewayAttachment = iga
							return CreateRouteTable(name+"-vpc-route-table", vpc).GetAndThen(ctx, func(rt *ec2.RouteTable) error {
								bnc.RouteTable = rt
								return CreateAndAttachDefaultRoute(name+"-route-default", rt, igw, iga).GetAndThen(ctx, func(r *ec2.Route) error {
									bnc.DefaultRoute = r
									return nil
								})
							})
						})
					}); err != nil {
						return err
					}
					if err := CreateSecurityGroup(name+"-security-group", vpc).GetAndThen(ctx, func(sg *ec2.SecurityGroup) error {
						bnc.LBSecurityGroup = sg
						return pgocomp.ApplyAll(ctx,
							CreateAndAttachTCPIngressSecurityGroupRule(name+"security-group-rule-https", sg, 443, 443, []string{"0.0.0.0/0"}),
						)
					}); err != nil {
						return err
					}
					return nil
				})
			}); err != nil {
				return nil, err
			}
			return &bnc, nil
		},
	)
}
