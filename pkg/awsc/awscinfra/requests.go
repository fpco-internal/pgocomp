package awscinfra

import (
	"github.com/fpco-internal/pgocomp"
)

// InfraParameters is a configuration for a
type InfraParameters struct {
	pgocomp.Meta
	Vpcs []VpcParameters
}

// ProviderParameters are used to create new region infrastructure
type ProviderParameters struct {
	pgocomp.Meta
	Region string
}

// VpcParameters are the parameters of the function NewBasicNetworkComponent
type VpcParameters struct {
	pgocomp.Meta
	Provider     ProviderParameters
	CidrBlock    string
	Partitions   []NetworkPartitionParameters
	Certificates []CertificateParameters
}

// ECSClusterParameters are parameters to create an ECS Cluster
type ECSClusterParameters struct {
	pgocomp.Meta
	Services []ECSServiceParameters
}

// ECSServiceParameters defines containers to be used in the Infra
type ECSServiceParameters struct {
	pgocomp.Meta
	DesiredCount   int
	CPU            int
	Memory         int
	AssignPublicIP bool
	Containers     []ContainerDefinition
}

// NetworkPartitionParameters defines the parameter for a network block, like public or private
type NetworkPartitionParameters struct {
	pgocomp.Meta
	IsPublic       bool
	Subnets        []SubnetParameters
	LoadBalancers  []LoadBalancerParameters
	LBTargetGroups []LBTargetGroupParameters
	ECSClusters    []ECSClusterParameters
}

// ContainerDefinition defines containers to be used in the Infra
type ContainerDefinition struct {
	Name         string                    `json:"name"`
	Image        string                    `json:"image"`
	Essential    *bool                     `json:"essential,omitempty"`
	PortMappings []ContainerPortMapping    `json:"portMappings,omitempty"`
	Environment  []ContainerEnvironmentVar `json:"environment,omitempty"`
	EntryPoint   []string                  `json:"entryPoint,omitempty"`
	Command      []string                  `json:"command,omitempty"`
	CPU          int64                     `json:"cpu,omitempty"`
	Memory       int64                     `json:"memory,omitempty"`
}

// ContainerEnvironmentVar ...
type ContainerEnvironmentVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ContainerPortMapping are the ports the the container exposes
type ContainerPortMapping struct {
	ContainerPort         int        `json:"containerPort"`
	HostPort              int        `json:"hostPort"`
	Protocol              TGProtocol `json:"protocol"`
	TargetGroupLookupName string     `json:"-"`
}

// SubnetParameters are parameters used by the CreateSubnet function
type SubnetParameters struct {
	pgocomp.Meta
	CidrBlock string
}

// LBType is the type of the network load balander
type LBType string

// LBProtocol is the network protocol tobe balanced
type LBProtocol string

const (
	//HTTP is the Hyper Text Protocol
	HTTP LBProtocol = "HTTP"
	//HTTP2 is the Hyper Text Protocol version 2
	HTTP2 LBProtocol = "HTTP2"
	//GRPC is the Google remote procedure call procotol used to change Protobuf messages
	GRPC LBProtocol = "GRPC"
)

// TGProtocol is the network protocol tobe balanced
type TGProtocol string

const (
	//TGProtoHTTP ...
	TGProtoHTTP TGProtocol = "HTTP"
	//TGProtoUDP ...
	TGProtoUDP TGProtocol = "UDP"
	//TGProtoTCP ...
	TGProtoTCP TGProtocol = "TCP"
	//TGProtoTCPUDP ...
	TGProtoTCPUDP TGProtocol = "TCP_UDP"
	//TGProtoHTTPS ...
	TGProtoHTTPS TGProtocol = "HTTPS"
	//TGProtoTLS ...
	TGProtoTLS TGProtocol = "TLS"
)

// LoadBalancerParameters are parameters used by the CreateSubnet function
type LoadBalancerParameters struct {
	pgocomp.Meta
	Type      LBType
	Listeners []LBListenerParameters
}

// LBRuleParameters is the paramenters for a listener
type LBRuleParameters struct {
	pgocomp.Meta
	Conditions []struct {
		Field  string
		values []string
	}
	ListenerLookupName    string
	TargetGroupLookupName string
	Priority              int
}

// LBListenerParameters is the paramenters for a listener
type LBListenerParameters struct {
	pgocomp.Meta
	Port                  int
	Protocol              LBProtocol
	TargetGroupLookupName string
	Rules                 []LBRuleParameters
	CertificateLookupName string
}

// CertificateValidationMethod is the method used to validate the certificate. By DNS or EMAIL
type CertificateValidationMethod string

const (
	//ValidationByDNS validates the certificate by a CNAME field in the DNS server. auto renew
	ValidationByDNS CertificateValidationMethod = "DNS"
	//ValidationByEmail validates by email.. manually renewed
	ValidationByEmail CertificateValidationMethod = "EMAIL"
)

// CertificateParameters creates a new certificate
type CertificateParameters struct {
	pgocomp.Meta
	Domain           string
	ValidationMethod CertificateValidationMethod
}

// TargetType the Target type of a Target group
type TargetType string

const (
	//TGIp is a target group that will group members by their ip adresses
	TGIp TargetType = "ip"
	//TGInstance is a target group that will group members by their instances
	TGInstance TargetType = "instance"
)

// LBTargetGroupParameters is the paramenters for a listener
type LBTargetGroupParameters struct {
	pgocomp.Meta
	Port       int
	Protocol   TGProtocol
	TargetType TargetType
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
