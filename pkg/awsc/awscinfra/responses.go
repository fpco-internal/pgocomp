package awscinfra

import (
	"github.com/fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
)

// SingleRegionInfra is the return type of the function NewBasicNetworkComponent
type SingleRegionInfra struct {
	Region *pgocomp.GetComponentWithMetaResponse[*VpcComponent]
}

// InfraComponent b
type InfraComponent struct {
	Vpcs map[string]*pgocomp.GetComponentWithMetaResponse[*VpcComponent]
}

// VpcComponent is the return type of the function NewBasicNetworkComponent
type VpcComponent struct {
	Provider *pgocomp.GetComponentWithMetaResponse[*aws.Provider]
	Vpc      *pgocomp.GetComponentWithMetaResponse[*ec2.Vpc]
	Gateway  struct {
		InternetGateway      *pgocomp.GetComponentWithMetaResponse[*ec2.InternetGateway]
		VpcGatewayAttachment *pgocomp.GetComponentWithMetaResponse[*ec2.InternetGatewayAttachment]
		RouteTable           *pgocomp.GetComponentWithMetaResponse[*ec2.RouteTable]
		DefaultRoute         *pgocomp.GetComponentWithMetaResponse[*ec2.Route]
	}
	Partitions   map[string]*pgocomp.GetComponentWithMetaResponse[*NetworkPartitionComponent]
	Certificates map[string]*pgocomp.GetComponentWithMetaResponse[*acm.Certificate]
}

// NetworkPartitionComponent is the response of CreateBlockComponent
type NetworkPartitionComponent struct {
	Subnets       map[string]*pgocomp.GetComponentWithMetaResponse[*ec2.Subnet]
	LoadBalancers map[string]*pgocomp.GetComponentWithMetaResponse[*LoadBalancerComponent]
	TargetGroups  map[string]*pgocomp.GetComponentWithMetaResponse[*lb.TargetGroup]
	ECSClusters   map[string]*pgocomp.GetComponentWithMetaResponse[*ECSClusterComponent]
}

// LoadBalancerComponent is the response of CreateLoadBalancerComponent function
type LoadBalancerComponent struct {
	LoadBalancer  *pgocomp.GetComponentWithMetaResponse[*lb.LoadBalancer]
	SecurityGroup *pgocomp.GetComponentWithMetaResponse[*ec2.SecurityGroup]
	Listeners     map[string]*pgocomp.GetComponentWithMetaResponse[*lb.Listener]
}

// ECSClusterComponent holds the created cluster components
type ECSClusterComponent struct {
	Cluster         *pgocomp.GetComponentWithMetaResponse[*ecs.Cluster]
	FargateServices map[string]*pgocomp.GetComponentWithMetaResponse[*ecs.Service]
}
