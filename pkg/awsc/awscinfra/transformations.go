package awscinfra

import (
	jsoniter "github.com/json-iterator/go"
	ecsn "github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ECSNativeTaskDefinitionContainerDefinitionArray transform containers into a ECS Native array
func (c *ECSServiceParameters) ECSNativeTaskDefinitionContainerDefinitionArray() (array ecsn.TaskDefinitionContainerDefinitionArray) {
	for _, container := range c.Containers {
		array = append(array, container.ECSNativeTaskDefinitionContainerDefinitionArgs())
	}
	return
}

// ECSNativeTaskDefinitionPortMappingArray transforms this configuration into a ECS Native Port Mapping Array
func (c *ContainerDefinition) ECSNativeTaskDefinitionPortMappingArray() (array ecsn.TaskDefinitionPortMappingArray) {
	for _, port := range c.PortMappings {
		array = append(array, ecsn.TaskDefinitionPortMappingArgs{
			ContainerPort: pulumi.Int(port.ContainerPort),
			AppProtocol:   port.ECSNativeTaskDefinitionPortMappingAppProtocol(),
		})
	}
	return
}

// ECSNativeTaskDefinitionPortMappingArgs transforms this configuration into a ECS Native Port Mapping Args
func (p *ContainerPortMapping) ECSNativeTaskDefinitionPortMappingArgs() (array ecsn.TaskDefinitionPortMappingArgs) {
	return ecsn.TaskDefinitionPortMappingArgs{
		ContainerPort: pulumi.Int(p.ContainerPort),
		AppProtocol:   p.ECSNativeTaskDefinitionPortMappingAppProtocol(),
	}
}

// ECSNativeTaskDefinitionPortMappingAppProtocol transforms ContainerPortMapping into a ecs native TaskDefinitionPortMappings
func (p *ContainerPortMapping) ECSNativeTaskDefinitionPortMappingAppProtocol() ecsn.TaskDefinitionPortMappingAppProtocol {
	if matchOrPanic("(?i)http2", string(p.Protocol)) {
		return ecsn.TaskDefinitionPortMappingAppProtocolHttp2
	} else if matchOrPanic("(?i)grpc", string(p.Protocol)) {
		return ecsn.TaskDefinitionPortMappingAppProtocolGrpc
	} else {
		return ecsn.TaskDefinitionPortMappingAppProtocolHttp
	}
}

// ECSNativeTaskDefinitionKeyValuePairArray transforms this container into a ECS Native KeyValue Pair Array
func (c *ContainerDefinition) ECSNativeTaskDefinitionKeyValuePairArray() (array ecsn.TaskDefinitionKeyValuePairArray) {
	for _, item := range c.Environment {
		array = append(array, ecsn.TaskDefinitionKeyValuePairArgs{
			Name:  pulumi.String(item.Name),
			Value: pulumi.String(item.Value),
		})
	}
	return
}

// ECSNativeTaskDefinitionContainerDefinitionArgs transformas a container parameter into a Ecs native container defitinion
func (c *ContainerDefinition) ECSNativeTaskDefinitionContainerDefinitionArgs() ecsn.TaskDefinitionContainerDefinitionArgs {
	return ecsn.TaskDefinitionContainerDefinitionArgs{
		Name:         pulumi.String(c.Name),
		Cpu:          pulumi.Int(c.CPU),
		Memory:       pulumi.Int(c.Memory),
		Image:        pulumi.String(c.Image),
		Environment:  c.ECSNativeTaskDefinitionKeyValuePairArray(),
		PortMappings: c.ECSNativeTaskDefinitionPortMappingArray(),
	}
}

// ECSTaskDefinitionContainerDefinitionArray transform containers into a json definition
func (c *ECSServiceParameters) ECSTaskDefinitionContainerDefinitionArray() (string, error) {
	jsonData, err := json.Marshal(c.Containers)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
