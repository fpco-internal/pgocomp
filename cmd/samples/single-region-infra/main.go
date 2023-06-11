package main

import (
	"os"

	"github.com/fpco-internal/pgocomp/pkg/awsc/awscinfra"

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
								Name:    "bots",
								Services: []awscinfra.ECSServiceParameters{
									{
										Element: awscinfra.Element{
											Active: true,
											Tags: map[string]string{
												"tagname": "tagvalue",
											},
										},
										Name:           "discord",
										DesiredCount:   1,
										CPU:            256,
										Memory:         512,
										AssignPublicIP: true,
										Containers: []awscinfra.ContainerParameters{
											{
												Name:   "discordbot",
												Image:  "gracig/bot:latest",
												CPU:    128,
												Memory: 256,
												Environment: map[string]string{
													"DISCORD_TOKEN": os.Getenv("DISCORD_TOKEN"),
												},
											},
										},
									},
									{
										Element: awscinfra.Element{
											Active: true,
											Tags: map[string]string{
												"tagname": "tagvalue",
											},
										},
										Name:           "simple-web",
										DesiredCount:   2,
										CPU:            256,
										Memory:         512,
										AssignPublicIP: true,
										Containers: []awscinfra.ContainerParameters{
											{
												Name:   "web80",
												Image:  "yeasy/simple-web:latest",
												CPU:    128,
												Memory: 256,
												PortMappings: []awscinfra.ContainerPortMapping{{
													ContainerPort:         80,
													TargetGroupLookupName: "web80",
												}},
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
						SubnetB: awscinfra.SubnetParameters{Element: awscinfra.Element{Active: true, Tags: map[string]string{"tagname": "tagvalue"}},
							AvailabilityZone: "us-east-2b",
							CidrBlock:        "10.0.6.0/24",
						},
					},
				},
			}).Apply(ctx)
	})
}
