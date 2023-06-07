package awscinfra

import (
	"fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
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
	DesiredCount   int
	CPU            string
	Memory         string
	AssignPublicIP bool
	LaunchType     ECSLaunchType
	Containers     []ContainerParameters
}

// GetPublicIP transforms a bool into a PublicIp value
func (p *ECSServiceParameters) GetPublicIP() ecs.ServiceAwsVpcConfigurationAssignPublicIp {
	var publicIP = ecs.ServiceAwsVpcConfigurationAssignPublicIpEnabled
	if !p.AssignPublicIP {
		publicIP = ecs.ServiceAwsVpcConfigurationAssignPublicIpDisabled
	}
	return publicIP
}

// NetworkPartitionParameters defines the parameter for a network block, like public or private
type NetworkPartitionParameters struct {
	SubnetA      SubnetParameters
	SubnetB      SubnetParameters
	LoadBalancer LoadBalancerParameters
	ECSClusters  []ECSClusterParameters
}

// ECSLaunchType set the type of the host that will run the container.
type ECSLaunchType string

const (
	//FargateLaunchType runs the container under the Fargate infrastructure
	FargateLaunchType ECSLaunchType = "FARGATE"
	//FargateSpotLaunchType runs the container under the Fargate infrastructure
	FargateSpotLaunchType ECSLaunchType = "FARGATE_SPOT"
	//EC2LaunchType runs the container under an ECS instance
	EC2LaunchType ECSLaunchType = "EC2"
)

// ContainerParameters defines containers to be used in the Infra
type ContainerParameters struct {
	//Name is the name of the container. It will ovewrite the Definition.Name attribute
	Name             string
	Definition       ecs.TaskDefinitionContainerDefinitionArgs
	LoadBalancerInfo []ContainerLBInfo
}

// ContainerLBInfo contains information for service load balancing of a container
type ContainerLBInfo struct {
	ContainerPort         int
	Protocol              LBProtocol
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
	Cluster  *pgocomp.GetComponentResponse[*ecs.Cluster]
	Services *pgocomp.GetComponentResponse[[]*ecs.Service]
}
