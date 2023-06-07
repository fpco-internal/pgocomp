package main

import (
	"os"

	"github.com/fpco-internal/pgocomp/pkg/awsc/awscinfra"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return awscinfra.CreateSingleRegionInfra(
			"my-infra",
			awscinfra.SingleRegionParameters{
				Region: awscinfra.RegionParameters{
					Element:   awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
					Region:    "us-east-2",
					CidrBlock: "10.0.0.0/16",
					Public: awscinfra.NetworkPartitionParameters{
						SubnetA: awscinfra.SubnetParameters{
							Element:          awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
							AvailabilityZone: "us-east-2a",
							CidrBlock:        "10.0.0.0/24",
						},
						SubnetB: awscinfra.SubnetParameters{
							Element:          awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
							AvailabilityZone: "us-east-2b",
							CidrBlock:        "10.0.1.0/24",
						},
						LoadBalancer: awscinfra.LoadBalancerParameters{
							Element: awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
							Type:    awscinfra.Application,
							Listeners: []awscinfra.LBListenerParameters{
								{
									Element:               awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
									Port:                  80,
									Protocol:              awscinfra.TCP,
									TargetGroupLookupName: "web80",
								},
							},
							TargetGroups: []awscinfra.LBTargetGroupParameters{
								{
									Element:    awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
									LookupName: "web80",
									Port:       80,
									Protocol:   awscinfra.TCP,
									TargetType: "ip",
								},
							},
						},
						ECSClusters: []awscinfra.ECSClusterParameters{
							{
								Element: awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
								Services: []awscinfra.ECSServiceParameters{
									{
										Element: awscinfra.Element{
											Active: true,
											Tags: map[string]string{
												"tagname": "tagvalue",
											}},
										DesiredCount:   1,
										CPU:            ".25 vCPU",
										Memory:         "512",
										AssignPublicIP: false,
										LaunchType:     awscinfra.FargateLaunchType,
										Containers: []awscinfra.ContainerParameters{
											{
												Name: "discordbot",
												Definition: ecs.TaskDefinitionContainerDefinitionArgs{
													Image:   pulumi.String("gracig/bot:latest"),
													Command: pulumi.ToStringArray([]string{"/discordbot"}),
													Environment: ecs.TaskDefinitionKeyValuePairArray{
														ecs.TaskDefinitionKeyValuePairArgs{
															Name:  pulumi.String("DISCORD_TOKEN"),
															Value: pulumi.String(os.Getenv("DISCORD_TOKEN")),
														},
													},
												},
											},
										},
									},
									{
										Element: awscinfra.Element{
											Active: false,
											Tags: map[string]string{
												"tagname": "tagvalue",
											}},
										DesiredCount:   1,
										CPU:            ".25 vCPU",
										Memory:         "512",
										AssignPublicIP: true,
										LaunchType:     awscinfra.FargateLaunchType,
										Containers: []awscinfra.ContainerParameters{
											{
												Name: "web80",
												Definition: ecs.TaskDefinitionContainerDefinitionArgs{
													Image: pulumi.String("yeasy/simple-web:latest"),
													PortMappings: ecs.TaskDefinitionPortMappingArray{
														ecs.TaskDefinitionPortMappingArgs{
															AppProtocol:   ecs.TaskDefinitionPortMappingAppProtocolHttp,
															HostPort:      pulumi.Int(80),
															ContainerPort: pulumi.Int(80),
															Protocol:      pulumi.String("tcp"),
														},
													},
												},
												LoadBalancerInfo: []awscinfra.ContainerLBInfo{
													{
														ContainerPort:         80,
														Protocol:              awscinfra.TCP,
														TargetGroupLookupName: "web80",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Private: awscinfra.NetworkPartitionParameters{
						SubnetA: awscinfra.SubnetParameters{
							Element:          awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
							AvailabilityZone: "us-east-2a",
							CidrBlock:        "10.0.5.0/24",
						},
						SubnetB: awscinfra.SubnetParameters{
							Element: awscinfra.Element{
								Active: true,
								Tags: map[string]string{
									"tagname": "tagvalue",
								}},
							AvailabilityZone: "us-east-2b",
							CidrBlock:        "10.0.6.0/24",
						},
					},
				},
			}).Apply(ctx)
	})
}
