package awsc

import (
	"github.com/fpco-internal/pgocomp"

	ecsn "github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"
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
func NewProvider(meta pgocomp.Meta, args *aws.ProviderArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*aws.Provider] {
	return pgocomp.NewPulumiComponentWithMeta(aws.NewProvider, meta, args, opts...)
}

// NewVpc is a wrapper to the ec2.NewVpc function
func NewVpc(meta pgocomp.Meta, args *ec2.VpcArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.Vpc] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewVpc, meta, args, opts...)
}

// NewListener is a wrapper to the lb.NewListener
func NewListener(meta pgocomp.Meta, args *lb.ListenerArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*lb.Listener] {
	return pgocomp.NewPulumiComponentWithMeta(lb.NewListener, meta, args, opts...)
}

// NewLoadBalancer is a wrapper to the lb.NewLoadBalancer
func NewLoadBalancer(meta pgocomp.Meta, args *lb.LoadBalancerArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*lb.LoadBalancer] {
	return pgocomp.NewPulumiComponentWithMeta(lb.NewLoadBalancer, meta, args, opts...)
}

// NewTargetGroup is a wrapper to the lb.NewTargetGroup
func NewTargetGroup(meta pgocomp.Meta, args *lb.TargetGroupArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*lb.TargetGroup] {
	return pgocomp.NewPulumiComponentWithMeta(lb.NewTargetGroup, meta, args, opts...)
}

// NewSecurityGroup is a wrapper to the ec2.NewSecurityGroup
func NewSecurityGroup(meta pgocomp.Meta, args *ec2.SecurityGroupArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.SecurityGroup] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewSecurityGroup, meta, args, opts...)
}

// NewSecurityGroupRule is a wrapper to the ec2.NewSecurityGroupRule function
func NewSecurityGroupRule(meta pgocomp.Meta, args *ec2.SecurityGroupRuleArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.SecurityGroupRule] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewSecurityGroupRule, meta, args, opts...)
}

// NewSubnet is a wrapper to the ec2.NewSubnet function
func NewSubnet(meta pgocomp.Meta, args *ec2.SubnetArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.Subnet] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewSubnet, meta, args, opts...)
}

// NewInternetGateway is a wrapper to the ec2.NewInternetGateway function
func NewInternetGateway(meta pgocomp.Meta, args *ec2.InternetGatewayArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.InternetGateway] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewInternetGateway, meta, args, opts...)
}

// NewInternetGatewayAttachment is a wrapper to the ec2.NewInternetGatewayAttachment function
func NewInternetGatewayAttachment(meta pgocomp.Meta, args *ec2.InternetGatewayAttachmentArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.InternetGatewayAttachment] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewInternetGatewayAttachment, meta, args, opts...)
}

// NewRoute is a wrapper to the ec2.NewRoute function
func NewRoute(meta pgocomp.Meta, args *ec2.RouteArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.Route] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewRoute, meta, args, opts...)
}

// NewRouteTable is a wrapper to the ec2.NewRouteTable function
func NewRouteTable(meta pgocomp.Meta, args *ec2.RouteTableArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.RouteTable] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewRouteTable, meta, args, opts...)
}

// NewCluster is a wrapper to the ec2.NewCluster function
func NewCluster(meta pgocomp.Meta, args *ecs.ClusterArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecs.Cluster] {
	return pgocomp.NewPulumiComponentWithMeta(ecs.NewCluster, meta, args, opts...)
}

// NewCapacityProvider is a wrapper to the ec2.NewCapacityProvider function
func NewCapacityProvider(meta pgocomp.Meta, args *ecs.CapacityProviderArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecs.CapacityProvider] {
	return pgocomp.NewPulumiComponentWithMeta(ecs.NewCapacityProvider, meta, args, opts...)
}

// NewECSService is a wrapper to the ec2.NewService function
func NewECSService(meta pgocomp.Meta, args *ecs.ServiceArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecs.Service] {
	return pgocomp.NewPulumiComponentWithMeta(ecs.NewService, meta, args, opts...)
}

// NewECSNativeService is a wrapper to the ecsn.NewECSNativeService function
func NewECSNativeService(meta pgocomp.Meta, args *ecsn.ServiceArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecsn.Service] {
	return pgocomp.NewPulumiComponentWithMeta(ecsn.NewService, meta, args, opts...)
}

// NewFargateService is a wrapper to the ec2.NewService function
func NewFargateService(meta pgocomp.Meta, args *ecsx.FargateServiceArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecsx.FargateService] {
	return pgocomp.NewPulumiComponentWithMeta(ecsx.NewFargateService, meta, args, opts...)
}

// NewEcsNativeTaskDefinition is a wrapper to the ec2.NewService function
func NewEcsNativeTaskDefinition(meta pgocomp.Meta, args *ecsn.TaskDefinitionArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecsn.TaskDefinition] {
	return pgocomp.NewPulumiComponentWithMeta(ecsn.NewTaskDefinition, meta, args, opts...)
}

// NewEcsTaskDefinition is a wrapper to the ec2.NewService function
func NewEcsTaskDefinition(meta pgocomp.Meta, args *ecs.TaskDefinitionArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ecs.TaskDefinition] {
	return pgocomp.NewPulumiComponentWithMeta(ecs.NewTaskDefinition, meta, args, opts...)
}

// NewRouteTableAssociation is a wrapper to the ec2.NewRouteTableAssociation function
func NewRouteTableAssociation(meta pgocomp.Meta, args *ec2.RouteTableAssociationArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*ec2.RouteTableAssociation] {
	return pgocomp.NewPulumiComponentWithMeta(ec2.NewRouteTableAssociation, meta, args, opts...)
}

// NewCertificate is a wrapped to create a new Acm certificate
func NewCertificate(meta pgocomp.Meta, args *acm.CertificateArgs, opts ...pulumi.ResourceOption) *pgocomp.ComponentWithMeta[*acm.Certificate] {
	return pgocomp.NewPulumiComponentWithMeta(acm.NewCertificate, meta, args, opts...)
}
