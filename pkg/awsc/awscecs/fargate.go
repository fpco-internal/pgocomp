package awscecs

import (
	"github.com/fpco-internal/pgocomp/pkg/awsc"

	"github.com/fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ContainerDefinition describes a container and also the dependencies under it
type ContainerDefinition struct {
	ContainerDefinitionArgs *ecs.TaskDefinitionContainerDefinitionArgs
	LoadBalancePorts        []int
	DependsOn               []pulumi.Resource
}

func CreateLBContainerDefinition() *pgocomp.Component[*ContainerDefinition] {
	return nil
}

// CreateFargateServiceParams as the parameters of the CreateFargateService function
type CreateFargateServiceParams struct {
	CPU                  string
	Memory               string
	DesiredCount         int
	Containers           []ContainerDefinition
	NetworkConfiguration *ecs.ServiceNetworkConfigurationArgs
}

// CreateFargateService takes a name, a list of container definition Functions
func CreateFargateService(
	name string,
	params CreateFargateServiceParams,
) *pgocomp.Component[*ecs.Service] {

	return awsc.NewLazyArgsService(name, func(ctx *pulumi.Context) (*ecs.ServiceArgs, []pulumi.ResourceOption, error) {

		//Creating the Task Definition
		var containerDefinitions ecs.TaskDefinitionContainerDefinitionArray
		var loadBalancerArgs []*ecs.ServiceLoadBalancerArgs
		var containerDeps []pulumi.Resource
		for _, cd := range params.Containers {
			if cd.ContainerDefinitionArgs != nil {
				containerDefinitions = append(containerDefinitions, cd.ContainerDefinitionArgs)
			}
			containerDeps = append(containerDeps, cd.DependsOn...)
			for _, port := range cd.LoadBalancePorts {
				loadBalancerArgs = append(loadBalancerArgs, &ecs.ServiceLoadBalancerArgs{
					ContainerName: cd.ContainerDefinitionArgs.Name,
					ContainerPort: pulumi.Int(port),
				})
			}
		}
		taskDef, err := ecs.NewTaskDefinition(ctx, name+"taskDef", &ecs.TaskDefinitionArgs{
			RequiresCompatibilities: pulumi.ToStringArray([]string{"FARGATE"}),
			Cpu:                     pulumi.String(params.CPU),
			Memory:                  pulumi.String(params.Memory),
			ContainerDefinitions:    containerDefinitions,
		}, pulumi.DependsOn(containerDeps))
		if err != nil {
			return nil, nil, err
		}

		return &ecs.ServiceArgs{
			LaunchType: ecs.ServiceLaunchTypeFargate,
			NetworkConfiguration: ecs.ServiceNetworkConfigurationArgs{
				AwsvpcConfiguration: ecs.ServiceAwsVpcConfigurationArgs{
					AssignPublicIp: ecs.ServiceAwsVpcConfigurationAssignPublicIpEnabled,
				},
			},
			TaskDefinition: taskDef.ID(),
		}, []pulumi.ResourceOption{pulumi.DependsOn([]pulumi.Resource{taskDef})}, nil
	})

}

/*
	return awsc.NewLazyArgsService(name, func() (*ecs.ServiceArgs, error) {
		//Create the service Task Definition Arguments
		taskDefinitionArgs := &ecs.TaskDefinitionArgs{
			RequiresCompatibilities: pulumi.ToStringArray([]string{"FARGATE"}),
			NetworkMode:             pulumi.String("awsvpc"),
			Cpu:                     pulumi.String(params.CPU),
			Memory:                  pulumi.String(params.Memory),
		}
		//Collect information from the CreateContainerFunctions
		var containers ecs.TaskDefinitionContainerDefinitionArray
		var loadBalancers ecs.ServiceLoadBalancerArray
		var opts []pulumi.ResourceOption
		for _, result := range params.ContainerDefinitionFunctions {
			if result, err := createContainer(); err != nil {
				if result.ContainerDefinitionArgs != nil {
					containers = append(containers, *result.ContainerDefinitionArgs)
				}
				if result.LoadBalancerArgs != nil {
					loadBalancers = append(loadBalancers, *result.LoadBalancerArgs)
				}
				opts = append(opts, pulumi.DependsOn(result.Dependencies))
			}
			return nil, err
		}
		//Set ContainerDefinition if any
		if len(containers) > 0 {
			taskDefinitionArgs.ContainerDefinitions = containers
		}

		return &ecs.ServiceArgs{
			LaunchType: ecs.ServiceLaunchTypeFargate,
			NetworkConfiguration: ecs.ServiceNetworkConfigurationArgs{
				AwsvpcConfiguration: ecs.ServiceAwsVpcConfigurationArgs{
					AssignPublicIp: ecs.ServiceAwsVpcConfigurationAssignPublicIpEnabled,
				},
			},
		}, nil

	})
*/
