package awsc

import (
	"fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewProvider is a wrapper to the aws.NewProvider function that returns a component of it
func NewProvider(name string, args *aws.ProviderArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*aws.Provider] {
	return pgocomp.NewPulumiComponent(aws.NewProvider, name, args, opts...)
}

// NewVpc is a wrapper to the ec2.NewVpc function that returns a component of it
func NewVpc(name string, args *ec2.VpcArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.Vpc] {
	return pgocomp.NewPulumiComponent(ec2.NewVpc, name, args, opts...)
}

// NewSecurityGroup is a wrapper to the ec2.NewSecurityGroup function that returns a component of it
func NewSecurityGroup(name string, args *ec2.SecurityGroupArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.SecurityGroup] {
	return pgocomp.NewPulumiComponent(ec2.NewSecurityGroup, name, args, opts...)
}

// NewSecurityGroupRule is a wrapper to the ec2.NewSecurityGroupRule function that returns a component of it
func NewSecurityGroupRule(name string, args *ec2.SecurityGroupRuleArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.SecurityGroupRule] {
	return pgocomp.NewPulumiComponent(ec2.NewSecurityGroupRule, name, args, opts...)
}

// NewSubnet is a wrapper to the ec2.NewSubnet function that returns a component of it
func NewSubnet(name string, args *ec2.SubnetArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.Subnet] {
	return pgocomp.NewPulumiComponent(ec2.NewSubnet, name, args, opts...)
}

// NewInternetGateway is a wrapper to the ec2.NewInternetGateway function that returns a component of it
func NewInternetGateway(name string, args *ec2.InternetGatewayArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.InternetGateway] {
	return pgocomp.NewPulumiComponent(ec2.NewInternetGateway, name, args, opts...)
}

// NewInternetGatewayAttachment is a wrapper to the ec2.NewInternetGatewayAttachment function that returns a component of it
func NewInternetGatewayAttachment(name string, args *ec2.InternetGatewayAttachmentArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.InternetGatewayAttachment] {
	return pgocomp.NewPulumiComponent(ec2.NewInternetGatewayAttachment, name, args, opts...)
}

// NewRoute is a wrapper to the ec2.NewRoute function that returns a component of it
func NewRoute(name string, args *ec2.RouteArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.Route] {
	return pgocomp.NewPulumiComponent(ec2.NewRoute, name, args, opts...)
}

// NewRouteTable is a wrapper to the ec2.NewRouteTable function that returns a component of it
func NewRouteTable(name string, args *ec2.RouteTableArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.RouteTable] {
	return pgocomp.NewPulumiComponent(ec2.NewRouteTable, name, args, opts...)
}
