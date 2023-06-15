package params

import (
	"github.com/fpco-internal/pgocomp"
)

// Region are the parameters of the function NewBasicNetworkComponent
type Region struct {
	pgocomp.Meta
	Region     string
	CidrBlock  string
	Partitions []NetworkPartitionParameters
}

// Cluster are parameters to create an ECS Cluster
type Cluster struct {
	pgocomp.Meta
	Services []Service
}

// Service defines containers to be used in the Infra
type Service struct {
	pgocomp.Meta
	Name           string
	DesiredCount   int
	CPU            int
	Memory         int
	AssignPublicIP bool
	Containers     []Container
}

// NetworkPartitionParameters defines the parameter for a network block, like public or private
type NetworkPartitionParameters struct {
	pgocomp.Meta
	Subnets        []Subnet
	LoadBalancers  []LoadBalancer
	LBTargetGroups []LBTargetGroup
	ECSClusters    []Cluster
}

// Container defines containers to be used in the Infra
type Container struct {
	Name         string
	Image        string
	CPU          int
	Memory       int
	PortMappings []ContainerPort
	Environment  map[string]string
}

// ContainerPort are the ports the the container exposes
type ContainerPort struct {
	ContainerPort     int
	AppProtocol       string
	TargetGroupLookup string
}

// VPC are parameters used by the CreateVPC function
type VPC struct {
	pgocomp.Meta
	CidrBlock string
}

// Subnet are parameters used by the CreateSubnet function
type Subnet struct {
	pgocomp.Meta
	CidrBlock        string
	AvailabilityZone string
	IsPublic         bool
}

// LBType is the type of the network load balander
type LBType string

// LBProtocol is the network protocol tobe balanced
type LBProtocol string

const (
	//TCP is the tcp protocol
	TCP LBProtocol = "TCP"
)

// LoadBalancer are parameters used by the CreateSubnet function
type LoadBalancer struct {
	pgocomp.Meta
	Type      LBType
	Listeners []LBListener
}

// LBListener is the paramenters for a listener
type LBListener struct {
	pgocomp.Meta
	Port                  int
	Protocol              LBProtocol
	TargetGroupLookupName string
}

// LBTargetGroup is the paramenters for a listener
type LBTargetGroup struct {
	pgocomp.Meta
	Port       int
	Protocol   LBProtocol
	TargetType string
}

// GatewayParameters are parameters used by the CreateSubnet function
type GatewayParameters struct {
	pgocomp.Meta
}

const (
	//Classic is the classic elastic load balancer that works on layer 4 and 7
	Classic LBType = "gateway"

	//Application is the application load balancer that works on layer 7
	Application LBType = "application"

	//Network is the network load balancer that works on layer 4
	Network LBType = "network"
)
