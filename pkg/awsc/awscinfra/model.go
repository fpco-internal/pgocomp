package awscinfra

import (
	"github.com/fpco-internal/pgocomp"

	ecsn "github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
)

// SingleRegionParameters is a configuration for a single region infra
type SingleRegionParameters struct {
	Region RegionParameters
}

// RegionParameters are the parameters of the function NewBasicNetworkComponent
type RegionParameters struct {
	Element
	Region    string
	CidrBlock string
	Public    NetworkPartitionParameters
	Private   NetworkPartitionParameters
}

// Element defines some basics aspect of a component
type Element struct {
	Active bool
	Tags   map[string]string
}

// ECSClusterParameters are parameters to create an ECS Cluster
type ECSClusterParameters struct {
	Element
	Name     string
	Services []ECSServiceParameters
}

// CapacityProviderParameters holds parameters for the total capacity of an ECS Cluster
type CapacityProviderParameters struct {
	TargetCapacity         int
	MinimumScalingStepSize float64
	MaximumScalingStepSize float64
}

// ECSServiceParameters defines containers to be used in the Infra
type ECSServiceParameters struct {
	Element
	Name           string
	DesiredCount   int
	CPU            int
	Memory         int
	AssignPublicIP bool
	Containers     []ContainerParameters
}

// NetworkPartitionParameters defines the parameter for a network block, like public or private
type NetworkPartitionParameters struct {
	SubnetA      SubnetParameters
	SubnetB      SubnetParameters
	LoadBalancer LoadBalancerParameters
	ECSClusters  []ECSClusterParameters
}

// ContainerParameters defines containers to be used in the Infra
type ContainerParameters struct {
	Name         string
	Image        string
	CPU          int
	Memory       int
	PortMappings []ContainerPortMapping
	Environment  map[string]string
}

// TaskDefinitionContainerDefinitionArray transform containers into a ECS Native array
func (c *ECSServiceParameters) TaskDefinitionContainerDefinitionArray() (array ecsn.TaskDefinitionContainerDefinitionArray) {
	for _, container := range c.Containers {
		array = append(array, container.TaskDefinitionContainerDefinitionArgs())
	}
	return
}

// TaskDefinitionPortMappingArray transforms this configuration into a ECS Native Port Mapping Array
func (c *ContainerParameters) TaskDefinitionPortMappingArray() (array ecsn.TaskDefinitionPortMappingArray) {
	for _, port := range c.PortMappings {
		array = append(array, ecsn.TaskDefinitionPortMappingArgs{
			ContainerPort: pulumi.Int(port.ContainerPort),
			AppProtocol:   port.TaskDefinitionPortMappingAppProtocol(),
		})
	}
	return
}

// TaskDefinitionPortMappingArgs transforms this configuration into a ECS Native Port Mapping Args
func (p *ContainerPortMapping) TaskDefinitionPortMappingArgs() (array ecsn.TaskDefinitionPortMappingArgs) {
	return ecsn.TaskDefinitionPortMappingArgs{
		ContainerPort: pulumi.Int(p.ContainerPort),
		AppProtocol:   p.TaskDefinitionPortMappingAppProtocol(),
	}
}

// TaskDefinitionPortMappingAppProtocol transforms ContainerPortMapping into a ecs native TaskDefinitionPortMappings
func (p *ContainerPortMapping) TaskDefinitionPortMappingAppProtocol() ecsn.TaskDefinitionPortMappingAppProtocol {
	if matchOrPanic("(?i)http2", p.AppProtocol) {
		return ecsn.TaskDefinitionPortMappingAppProtocolHttp2
	} else if matchOrPanic("(?i)grpc", p.AppProtocol) {
		return ecsn.TaskDefinitionPortMappingAppProtocolGrpc
	} else {
		return ecsn.TaskDefinitionPortMappingAppProtocolHttp
	}
}

// TaskDefinitionKeyValuePairArray transforms this container into a ECS Native KeyValue Pair Array
func (c *ContainerParameters) TaskDefinitionKeyValuePairArray() (array ecsn.TaskDefinitionKeyValuePairArray) {
	for key, value := range c.Environment {
		array = append(array, ecsn.TaskDefinitionKeyValuePairArgs{
			Name:  pulumi.String(key),
			Value: pulumi.String(value),
		})
	}
	return
}

// TaskDefinitionContainerDefinitionArgs transformas a container parameter into a Ecs native container defitinion
func (c *ContainerParameters) TaskDefinitionContainerDefinitionArgs() ecsn.TaskDefinitionContainerDefinitionArgs {
	return ecsn.TaskDefinitionContainerDefinitionArgs{
		Name:         pulumi.String(c.Name),
		Cpu:          pulumi.Int(c.CPU),
		Memory:       pulumi.Int(c.Memory),
		Image:        pulumi.String(c.Image),
		Environment:  c.TaskDefinitionKeyValuePairArray(),
		PortMappings: c.TaskDefinitionPortMappingArray(),
	}
}

// ContainerPortMapping are the ports the the container exposes
type ContainerPortMapping struct {
	ContainerPort         int
	AppProtocol           string //HTTP HTTPS GRPC
	TargetGroupLookupName string
}

// VPCParameters are parameters used by the CreateVPC function
type VPCParameters struct {
	Element
	CidrBlock string
}

// SubnetParameters are parameters used by the CreateSubnet function
type SubnetParameters struct {
	Element
	CidrBlock        string
	AvailabilityZone string
}

// LBType is the type of the network load balander
type LBType string

// LBProtocol is the network protocol tobe balanced
type LBProtocol string

const (
	//TCP is the tcp protocol
	TCP LBProtocol = "TCP"
)

// LoadBalancerParameters are parameters used by the CreateSubnet function
type LoadBalancerParameters struct {
	Element
	Type         LBType
	Listeners    []LBListenerParameters
	TargetGroups []LBTargetGroupParameters
}

// LBListenerParameters is the paramenters for a listener
type LBListenerParameters struct {
	Element
	Port                  int
	Protocol              LBProtocol
	TargetGroupLookupName string
}

// LBTargetGroupParameters is the paramenters for a listener
type LBTargetGroupParameters struct {
	Element
	LookupName string
	Port       int
	Protocol   LBProtocol
	TargetType string
}

// GatewayParameters are parameters used by the CreateSubnet function
type GatewayParameters struct {
	Element
}

const (
	//Classic is the classic elastic load balancer that works on layer 4 and 7
	Classic LBType = "gateway"

	//Application is the application load balancer that works on layer 7
	Application LBType = "application"

	//Network is the network load balancer that works on layer 4
	Network LBType = "network"
)

// SingleRegionInfra is the return type of the function NewBasicNetworkComponent
type SingleRegionInfra struct {
	Region *pgocomp.GetComponentResponse[*RegionComponent]
}

// RegionComponent is the return type of the function NewBasicNetworkComponent
type RegionComponent struct {
	Provider *pgocomp.GetComponentResponse[*aws.Provider]
	Vpc      *pgocomp.GetComponentResponse[*ec2.Vpc]
	Gateway  struct {
		InternetGateway      *pgocomp.GetComponentResponse[*ec2.InternetGateway]
		VpcGatewayAttachment *pgocomp.GetComponentResponse[*ec2.InternetGatewayAttachment]
		RouteTable           *pgocomp.GetComponentResponse[*ec2.RouteTable]
		DefaultRoute         *pgocomp.GetComponentResponse[*ec2.Route]
	}
	Partitions struct {
		Public  *pgocomp.GetComponentResponse[*NetworkPartitionComponent]
		Private *pgocomp.GetComponentResponse[*NetworkPartitionComponent]
	}
}

// NetworkPartitionComponent is the response of CreateBlockComponent
type NetworkPartitionComponent struct {
	SubnetA      *pgocomp.GetComponentResponse[*ec2.Subnet]
	SubnetB      *pgocomp.GetComponentResponse[*ec2.Subnet]
	LoadBalancer *pgocomp.GetComponentResponse[*LoadBalancerComponent]
	ECSClusters  *pgocomp.GetComponentResponse[[]*ECSClusterComponent]
}

// LoadBalancerComponent is the response of CreateLoadBalancerComponent function
type LoadBalancerComponent struct {
	LoadBalancer  *pgocomp.GetComponentResponse[*lb.LoadBalancer]
	SecurityGroup *pgocomp.GetComponentResponse[*ec2.SecurityGroup]
	listeners     *pgocomp.GetComponentResponse[[]*lb.Listener]
	targetGroups  *pgocomp.GetComponentResponse[map[string]*lb.TargetGroup]
}

// ECSClusterComponent holds the created cluster components
type ECSClusterComponent struct {
	Cluster         *pgocomp.GetComponentResponse[*ecs.Cluster]
	FargateServices *pgocomp.GetComponentResponse[[]*ecsn.Service]
}
