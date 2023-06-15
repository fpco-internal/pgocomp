package main

import (
	pgo "github.com/fpco-internal/pgocomp"
	"github.com/fpco-internal/pgocomp/pkg/awsc/awscinfra"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return awscinfra.New(awscinfra.InfraParameters{
			Meta: pgo.Meta{Name: "myinfra"},
			Vpcs: []awscinfra.VpcParameters{{
				Meta: pgo.Meta{Name: "myvpc"},
				Provider: awscinfra.ProviderParameters{
					Meta:   pgo.Meta{Name: "myprovider"},
					Region: "us-east-1",
				},
				CidrBlock: "10.0.0.0/16",
				Partitions: []awscinfra.NetworkPartitionParameters{{
					Meta:     pgo.Meta{Name: "public"},
					IsPublic: true,
					Subnets: []awscinfra.SubnetParameters{{
						Meta:      pgo.Meta{Name: "subnet1"},
						CidrBlock: "10.0.1.0/24",
					}, {
						Meta:      pgo.Meta{Name: "subnet2"},
						CidrBlock: "10.0.2.0/24",
					}},
					LoadBalancers: []awscinfra.LoadBalancerParameters{{
						Meta: pgo.Meta{Name: "lb1", Protect: false},
						Type: awscinfra.Application,
						Listeners: []awscinfra.LBListenerParameters{{
							Meta:                  pgo.Meta{Name: "http"},
							Port:                  80,
							Protocol:              awscinfra.HTTP,
							TargetGroupLookupName: "http",
						}, {
							Meta:                  pgo.Meta{Name: "http8080"},
							Port:                  8080,
							Protocol:              awscinfra.HTTP,
							TargetGroupLookupName: "http",
							Rules: []awscinfra.LBRuleParameters{
								{
									Meta:                  pgo.Meta{Name: "http8080-rule1"},
									TargetGroupLookupName: "http",
									Priority:              1,
									Conditions: []awscinfra.LBRuleConditionParameters{
										{
											RuleConditionType: awscinfra.PathPattern,
											PathPatterns:      []string{"/web1"},
										},
									},
								},
								{
									Meta:                  pgo.Meta{Name: "http8080-rule2"},
									TargetGroupLookupName: "http",
									Priority:              2,
									Conditions: []awscinfra.LBRuleConditionParameters{
										{
											RuleConditionType: awscinfra.PathPattern,
											PathPatterns:      []string{"/web2"},
										},
									},
								},
							},
						}},
					}},
					LBTargetGroups: []awscinfra.LBTargetGroupParameters{{
						Meta:       pgo.Meta{Name: "http"},
						Port:       80,
						Protocol:   awscinfra.TGProtoHTTP,
						TargetType: awscinfra.TGIp,
					},
					},
					ECSClusters: []awscinfra.ECSClusterParameters{{
						Meta: pgo.Meta{
							Name: "cluster1",
						},
						Services: []awscinfra.ECSServiceParameters{{
							Meta: pgo.Meta{
								Name: "discord",
							},
							DesiredCount:   1,
							CPU:            256,
							Memory:         512,
							AssignPublicIP: true,
							Containers: []awscinfra.ContainerDefinition{{
								Name:   "http",
								Image:  "yeasy/simple-web:latest",
								CPU:    128,
								Memory: 256,
								PortMappings: []awscinfra.ContainerPortMapping{{
									ContainerPort:         80,
									HostPort:              80,
									Protocol:              awscinfra.TGProtoHTTP,
									TargetGroupLookupName: "http",
								}}, //PortMappings
							}}, //Containers
						}}, //Services
					}}, //ECSClusters
				}}, //Partitions
			}}, //VPCS
		}, //Infra
		).Apply(ctx)
	})
}
