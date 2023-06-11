package awsc

import (
	"github.com/fpco-internal/pgocomp"

	ecsn "github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	ecsx "github.com/pulumi/pulumi-awsx/sdk/go/awsx/ecs"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ToIDStringArray is a helper that takes a variadic list of CustomResourceState and return its IDs
func ToIDStringArray[T pulumi.CustomResourceState](comps ...*pulumi.CustomResourceState) pulumi.StringArray {
	var ids pulumi.StringArray
	for _, comp := range comps {
		ids = append(ids, comp.ID())
	}
	return ids
}

// NewProvider is a wrapper to the aws.NewProvider function
func NewProvider(name string, args *aws.ProviderArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*aws.Provider] {
	return pgocomp.NewPulumiComponent(aws.NewProvider, name, args, opts...)
}

// NewVpc is a wrapper to the ec2.NewVpc function
func NewVpc(name string, args *ec2.VpcArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.Vpc] {
	return pgocomp.NewPulumiComponent(ec2.NewVpc, name, args, opts...)
}

// NewListener is a wrapper to the lb.NewListener
func NewListener(name string, args *lb.ListenerArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*lb.Listener] {
	return pgocomp.NewPulumiComponent(lb.NewListener, name, args, opts...)
}

// NewLoadBalancer is a wrapper to the lb.NewLoadBalancer
func NewLoadBalancer(name string, args *lb.LoadBalancerArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*lb.LoadBalancer] {
	return pgocomp.NewPulumiComponent(lb.NewLoadBalancer, name, args, opts...)
}

// NewTargetGroup is a wrapper to the lb.NewTargetGroup
func NewTargetGroup(name string, args *lb.TargetGroupArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*lb.TargetGroup] {
	return pgocomp.NewPulumiComponent(lb.NewTargetGroup, name, args, opts...)
}

// NewSecurityGroup is a wrapper to the ec2.NewSecurityGroup
func NewSecurityGroup(name string, args *ec2.SecurityGroupArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.SecurityGroup] {
	return pgocomp.NewPulumiComponent(ec2.NewSecurityGroup, name, args, opts...)
}

// NewSecurityGroupRule is a wrapper to the ec2.NewSecurityGroupRule function
func NewSecurityGroupRule(name string, args *ec2.SecurityGroupRuleArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.SecurityGroupRule] {
	return pgocomp.NewPulumiComponent(ec2.NewSecurityGroupRule, name, args, opts...)
}

// NewSubnet is a wrapper to the ec2.NewSubnet function
func NewSubnet(name string, args *ec2.SubnetArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.Subnet] {
	return pgocomp.NewPulumiComponent(ec2.NewSubnet, name, args, opts...)
}

// NewInternetGateway is a wrapper to the ec2.NewInternetGateway function
func NewInternetGateway(name string, args *ec2.InternetGatewayArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.InternetGateway] {
	return pgocomp.NewPulumiComponent(ec2.NewInternetGateway, name, args, opts...)
}

// NewInternetGatewayAttachment is a wrapper to the ec2.NewInternetGatewayAttachment function
func NewInternetGatewayAttachment(name string, args *ec2.InternetGatewayAttachmentArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.InternetGatewayAttachment] {
	return pgocomp.NewPulumiComponent(ec2.NewInternetGatewayAttachment, name, args, opts...)
}

// NewRoute is a wrapper to the ec2.NewRoute function
func NewRoute(name string, args *ec2.RouteArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.Route] {
	return pgocomp.NewPulumiComponent(ec2.NewRoute, name, args, opts...)
}

// NewRouteTable is a wrapper to the ec2.NewRouteTable function
func NewRouteTable(name string, args *ec2.RouteTableArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.RouteTable] {
	return pgocomp.NewPulumiComponent(ec2.NewRouteTable, name, args, opts...)
}

// NewCluster is a wrapper to the ec2.NewCluster function
func NewCluster(name string, args *ecs.ClusterArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ecs.Cluster] {
	return pgocomp.NewPulumiComponent(ecs.NewCluster, name, args, opts...)
}

// NewCapacityProvider is a wrapper to the ec2.NewCapacityProvider function
func NewCapacityProvider(name string, args *ecs.CapacityProviderArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ecs.CapacityProvider] {
	return pgocomp.NewPulumiComponent(ecs.NewCapacityProvider, name, args, opts...)
}

// NewService is a wrapper to the ec2.NewService function
func NewService(name string, args *ecs.ServiceArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ecs.Service] {
	return pgocomp.NewPulumiComponent(ecs.NewService, name, args, opts...)
}

// NewECSNativeService is a wrapper to the ecsn.NewECSNativeService function
func NewECSNativeService(name string, args *ecsn.ServiceArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ecsn.Service] {
	return pgocomp.NewPulumiComponent(ecsn.NewService, name, args, opts...)
}

// NewFargateService is a wrapper to the ec2.NewService function
func NewFargateService(name string, args *ecsx.FargateServiceArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ecsx.FargateService] {
	return pgocomp.NewPulumiComponent(ecsx.NewFargateService, name, args, opts...)
}

// NewEcsNativeTaskDefinition is a wrapper to the ec2.NewService function
func NewEcsNativeTaskDefinition(name string, args *ecsn.TaskDefinitionArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ecsn.TaskDefinition] {
	return pgocomp.NewPulumiComponent(ecsn.NewTaskDefinition, name, args, opts...)
}

// NewLazyArgsService is a wrapper to the ec2.NewService function
func NewLazyArgsService(name string, argsFn func(*pulumi.Context) (*ecs.ServiceArgs, []pulumi.ResourceOption, error)) *pgocomp.Component[*ecs.Service] {
	return pgocomp.NewLazyArgsPulumiComponent(ecs.NewService, name, argsFn)
}

// NewRouteTableAssociation is a wrapper to the ec2.NewRouteTableAssociation function
func NewRouteTableAssociation(name string, args *ec2.RouteTableAssociationArgs, opts ...pulumi.ResourceOption) *pgocomp.Component[*ec2.RouteTableAssociation] {
	return pgocomp.NewPulumiComponent(ec2.NewRouteTableAssociation, name, args, opts...)
}
