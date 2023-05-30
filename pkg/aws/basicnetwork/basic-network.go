package awsbasicnetwork

import (
	"fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BasicNetworkParameters struct {
	Vpc           VpcParameters
	PublicSubnet  SubnetParameters
	PrivateSubnet SubnetParameters
}

type VpcParameters struct {
	Region    string
	CidrBlock string
}

type SecurityGroupRuleTCPParams struct {
	FromPort   int
	ToPort     int
	CidrBlocks []string
}

type SubnetParameters struct {
	CidrBlock        string
	AvailabilityZone string
}

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

func BuildBasicNetwork(name string, params BasicNetworkParameters) *pgocomp.Component[*BasicNetwork] {
	return pgocomp.NewComponent(name,
		func(ctx *pulumi.Context) (*BasicNetwork, error) {
			var bnc BasicNetwork

			//The Provider
			if err := basicNetworkProvider(name+"-provider", params.Vpc.Region).GetAndThen(ctx, func(provider *aws.Provider) error {
				bnc.Provider = provider

				//The VPC
				return basicNetworkVpc(name+"-vpc", params.Vpc, provider).GetAndThen(ctx, func(vpc *ec2.Vpc) error {
					bnc.Vpc = vpc

					//Public Subnet
					if err := basicNetworkSubnet(name+"-public-subnet", params.PublicSubnet, vpc).GetAndThen(ctx, func(subnet *ec2.Subnet) error {
						bnc.PublicSubnet = subnet

						return nil
					}); err != nil {
						return err
					}

					//Private Subnet
					if err := basicNetworkSubnet(name+"-private-subnet", params.PrivateSubnet, vpc).GetAndThen(ctx, func(subnet *ec2.Subnet) error {
						bnc.PrivateSubnet = subnet

						return nil
					}); err != nil {
						return err
					}

					//Internet Gateway
					if err := basicNetworkIGW(name+"-igw", vpc).GetAndThen(ctx, func(igw *ec2.InternetGateway) error {
						bnc.IGW = igw

						//Attach  Gateway to the VPC
						return basicNetworkIGWAttachment(name+"-igw-vpc-attachment", vpc, igw).GetAndThen(ctx, func(iga *ec2.InternetGatewayAttachment) error {
							bnc.VpcGatewayAttachment = iga

							//Route Table
							return basicNetworkRouteTable(name+"-vpc-route-table", vpc).GetAndThen(ctx, func(rt *ec2.RouteTable) error {
								bnc.RouteTable = rt

								//Default Route for the route table
								return basicNetworkRouteDefault(name+"-route-default", rt, igw, iga).GetAndThen(ctx, func(r *ec2.Route) error {
									bnc.DefaultRoute = r

									return nil
								})
							})
						})

					}); err != nil {
						return err
					}

					if err := basicSecurityGroup(name+"-security-group", vpc).GetAndThen(ctx, func(sg *ec2.SecurityGroup) error {
						bnc.LBSecurityGroup = sg
						return pgocomp.ApplyAll(ctx,
							basicSecurityGroupRuleTCPIngress(name+"security-group-rule-https", sg, 443, 443, []string{"0.0.0.0/0"}),
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

func basicNetworkProvider(name string, region string) *pgocomp.Component[*aws.Provider] {
	return pgocomp.NewPulumiComponent(
		aws.NewProvider,
		name,
		&aws.ProviderArgs{
			Region: pulumi.String(region),
		})
}

func basicNetworkVpc(name string, params VpcParameters, provider *aws.Provider) *pgocomp.Component[*ec2.Vpc] {
	return pgocomp.NewPulumiComponent(
		ec2.NewVpc,
		name,
		&ec2.VpcArgs{
			CidrBlock:        pulumi.String(params.CidrBlock),
			EnableDnsSupport: pulumi.Bool(true),
		},
		pulumi.Provider(provider).(pulumi.ResourceOption),
	)
}

func basicSecurityGroup(name string, vpc *ec2.Vpc) *pgocomp.Component[*ec2.SecurityGroup] {
	return pgocomp.NewPulumiComponent(
		ec2.NewSecurityGroup,
		name,
		&ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

func basicSecurityGroupRuleTCPIngress(name string, sg *ec2.SecurityGroup, fromPort, toPort int, cidrBlocks []string) *pgocomp.Component[*ec2.SecurityGroupRule] {
	return basicSecurityGroupRule(name, &ec2.SecurityGroupRuleArgs{
		Type:            pulumi.String("ingress"),
		Protocol:        pulumi.String(("tcp")),
		SecurityGroupId: sg.ID(),
		CidrBlocks:      pulumi.ToStringArray(cidrBlocks),
		FromPort:        pulumi.Int(fromPort),
		ToPort:          pulumi.Int(toPort),
	}, pulumi.DependsOn([]pulumi.Resource{sg}))
}

func basicSecurityGroupRule(name string, args *ec2.SecurityGroupRuleArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.SecurityGroupRule] {
	return pgocomp.NewPulumiComponent(
		ec2.NewSecurityGroupRule,
		name,
		args,
		opts...,
	)
}

func basicNetworkSubnet(name string, params SubnetParameters, vpc *ec2.Vpc) *pgocomp.Component[*ec2.Subnet] {
	return pgocomp.NewPulumiComponent(
		ec2.NewSubnet,
		name,
		&ec2.SubnetArgs{
			AvailabilityZone: pulumi.String(params.AvailabilityZone),
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(params.CidrBlock),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

func basicNetworkIGW(name string, vpc *ec2.Vpc) *pgocomp.Component[*ec2.InternetGateway] {
	return pgocomp.NewPulumiComponent(
		ec2.NewInternetGateway,
		name,
		&ec2.InternetGatewayArgs{},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

func basicNetworkIGWAttachment(name string, vpc *ec2.Vpc, igw *ec2.InternetGateway) *pgocomp.Component[*ec2.InternetGatewayAttachment] {
	return pgocomp.NewPulumiComponent(
		ec2.NewInternetGatewayAttachment,
		name,
		&ec2.InternetGatewayAttachmentArgs{
			VpcId:             vpc.ID(),
			InternetGatewayId: igw.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc, igw}),
	)
}

func basicNetworkRouteTable(name string, vpc *ec2.Vpc) *pgocomp.Component[*ec2.RouteTable] {
	return pgocomp.NewPulumiComponent(
		ec2.NewRouteTable,
		name,
		&ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

func basicNetworkRouteDefault(name string, rt *ec2.RouteTable, igw *ec2.InternetGateway, iga *ec2.InternetGatewayAttachment) *pgocomp.Component[*ec2.Route] {
	return pgocomp.NewPulumiComponent(
		ec2.NewRoute,
		name,
		&ec2.RouteArgs{
			RouteTableId:         rt.ID(),
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			GatewayId:            igw.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{igw, iga}),
	)
}
