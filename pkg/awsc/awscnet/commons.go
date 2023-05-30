package awscnet

import (
	"fpco-internal/pgocomp"
	"fpco-internal/pgocomp/pkg/awsc"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateProvider takes a name and a region and returns an aws.Provider Component
func CreateProvider(name string, region string) *pgocomp.Component[*aws.Provider] {
	return awsc.NewProvider(name, &aws.ProviderArgs{Region: pulumi.String(region)})
}

// CreateVpcParameters are parameters used by the CreateVPC function
type CreateVpcParameters struct {
	Region    string
	CidrBlock string
}

// CreateVPC takes a name, a list of parameters and a provides and returns a Vpc Component
func CreateVPC(name string, params CreateVpcParameters, provider *aws.Provider) *pgocomp.Component[*ec2.Vpc] {
	return awsc.NewVpc(
		name,
		&ec2.VpcArgs{
			CidrBlock:        pulumi.String(params.CidrBlock),
			EnableDnsSupport: pulumi.Bool(true),
		},
		pulumi.Provider(provider).(pulumi.ResourceOption),
	)
}

// CreateSecurityGroup takes a name and a vpc and returns a SecurityGroup Component
func CreateSecurityGroup(name string, vpc *ec2.Vpc) *pgocomp.Component[*ec2.SecurityGroup] {
	return awsc.NewSecurityGroup(
		name,
		&ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateAndAttachTCPIngressSecurityGroupRule takes a name, a security group, some nework parameters and returns a SecurityGroupRule Component
func CreateAndAttachTCPIngressSecurityGroupRule(name string, sg *ec2.SecurityGroup, fromPort, toPort int, cidrBlocks []string) *pgocomp.Component[*ec2.SecurityGroupRule] {
	return awsc.NewSecurityGroupRule(
		name, &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			Protocol:        pulumi.String(("tcp")),
			SecurityGroupId: sg.ID(),
			CidrBlocks:      pulumi.ToStringArray(cidrBlocks),
			FromPort:        pulumi.Int(fromPort),
			ToPort:          pulumi.Int(toPort),
		}, pulumi.DependsOn([]pulumi.Resource{sg}))
}

// CreateSubnetParameters are parameters used by the CreateSubnet function
type CreateSubnetParameters struct {
	CidrBlock        string
	AvailabilityZone string
}

// CreateSubnet takes a name, some parameters and a vpc and returns a SubnetComponent
func CreateSubnet(name string, params CreateSubnetParameters, vpc *ec2.Vpc) *pgocomp.Component[*ec2.Subnet] {
	return awsc.NewSubnet(
		name,
		&ec2.SubnetArgs{
			AvailabilityZone: pulumi.String(params.AvailabilityZone),
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(params.CidrBlock),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateInternetGateway takes a name, a vpc and returns an InternetGateway component
func CreateInternetGateway(name string, vpc *ec2.Vpc) *pgocomp.Component[*ec2.InternetGateway] {
	return awsc.NewInternetGateway(
		name,
		&ec2.InternetGatewayArgs{},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// AttachInternetGatewayToVPC takes a name, a vpc, an internet gateway and returnas an InternetGatewayAttachment Component
func AttachInternetGatewayToVPC(name string, vpc *ec2.Vpc, igw *ec2.InternetGateway) *pgocomp.Component[*ec2.InternetGatewayAttachment] {
	return awsc.NewInternetGatewayAttachment(
		name,
		&ec2.InternetGatewayAttachmentArgs{
			VpcId:             vpc.ID(),
			InternetGatewayId: igw.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc, igw}),
	)
}

// CreateRouteTable takes a name and a vpc and returns a RouteTable Component
func CreateRouteTable(name string, vpc *ec2.Vpc) *pgocomp.Component[*ec2.RouteTable] {
	return awsc.NewRouteTable(
		name,
		&ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateAndAttachDefaultRoute takes a name, a route table, an internetgateway and an internet gateway attachment and returns a Route Component
func CreateAndAttachDefaultRoute(name string, rt *ec2.RouteTable, igw *ec2.InternetGateway, iga *ec2.InternetGatewayAttachment) *pgocomp.Component[*ec2.Route] {
	return awsc.NewRoute(
		name,
		&ec2.RouteArgs{
			RouteTableId:         rt.ID(),
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			GatewayId:            igw.ID(),
		},
		pulumi.DependsOn([]pulumi.Resource{igw, iga}),
	)
}
