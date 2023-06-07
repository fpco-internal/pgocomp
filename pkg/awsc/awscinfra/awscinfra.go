package awscinfra

import (
	"errors"
	"fpco-internal/pgocomp"
	"fpco-internal/pgocomp/pkg/awsc"
	"strconv"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//TODO implement TAGs on every pulumi Resource Create.. method

// CreateSingleRegionInfra creates a Region Component thar comprises of a Vpc and its subcomponents
func CreateSingleRegionInfra(name string, params SingleRegionParameters) *pgocomp.Component[*SingleRegionInfra] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*SingleRegionInfra, error) {
		var sgi SingleRegionInfra
		var err = CreateRegionComponent(name, params.Region).GetAndThen(ctx, func(region *pgocomp.GetComponentResponse[*RegionComponent]) error {
			sgi.Region = region
			return nil
		})
		return &sgi, err
	},
	)
}

// CreateRegionComponent creates a Region Component thar comprises of a Vpc and its subcomponents
func CreateRegionComponent(name string, params RegionParameters) *pgocomp.Component[*RegionComponent] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*RegionComponent, error) {
		var bnc RegionComponent //The result component that will hold all components
		var err = errors.Join(
			CreateProvider(name+"-provider", params.Region).GetAndThen(ctx, func(provider *pgocomp.GetComponentResponse[*aws.Provider]) error {
				bnc.Provider = pgocomp.ExportURN(ctx, provider)
				return CreateVPC(name+"-vpc", params, provider.Component).GetAndThen(ctx, func(vpc *pgocomp.GetComponentResponse[*ec2.Vpc]) error {
					bnc.Vpc = vpc
					return errors.Join(
						CreateInternetGateway(name+"-igw", provider.Component, vpc.Component).GetAndThen(ctx, func(igw *pgocomp.GetComponentResponse[*ec2.InternetGateway]) error {
							bnc.Gateway.InternetGateway = igw
							return AttachInternetGatewayToVPC(name+"-igw-vpc-attachment", provider.Component, vpc.Component, igw.Component).GetAndThen(ctx, func(iga *pgocomp.GetComponentResponse[*ec2.InternetGatewayAttachment]) error {
								bnc.Gateway.VpcGatewayAttachment = iga
								return CreateRouteTable(name+"-vpc-route-table", provider.Component, vpc.Component).GetAndThen(ctx, func(rt *pgocomp.GetComponentResponse[*ec2.RouteTable]) error {
									bnc.Gateway.RouteTable = rt
									return errors.Join(
										CreateAndAttachDefaultRoute(name+"-default-route", provider.Component, rt.Component, igw.Component, iga.Component).GetAndThen(ctx, func(r *pgocomp.GetComponentResponse[*ec2.Route]) error {
											bnc.Gateway.DefaultRoute = r
											return nil
										}),
									)
								})
							})
						}),
						CreatePublicNetworkPartition(name+"-public", params.Public, provider.Component, vpc.Component, bnc.Gateway.RouteTable.Component).GetAndThen(ctx, func(npc *pgocomp.GetComponentResponse[*NetworkPartitionComponent]) error {
							bnc.Partitions.Public = npc
							return nil
						}),
						CreatePrivateNetworkPartition(name+"-private", params.Private, provider.Component, vpc.Component).GetAndThen(ctx, func(npc *pgocomp.GetComponentResponse[*NetworkPartitionComponent]) error {
							bnc.Partitions.Private = npc
							return nil
						}),
					)
				})
			}),
		)
		return &bnc, err
	},
	)
}

// CreatePublicNetworkPartition takes some paramenters and creates a new Network Partition
func CreatePublicNetworkPartition(name string, params NetworkPartitionParameters, provider *aws.Provider, vpc *ec2.Vpc, rt *ec2.RouteTable) *pgocomp.Component[*NetworkPartitionComponent] {
	return CreateNetworkPartition(name, params, provider, vpc, rt)
}

// CreatePrivateNetworkPartition takes some paramenters and creates a new Network Partition
func CreatePrivateNetworkPartition(name string, params NetworkPartitionParameters, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.Component[*NetworkPartitionComponent] {
	return CreateNetworkPartition(name, params, provider, vpc, nil)
}

// CreateNetworkPartition takes some paramenters and creates a new Network Partition
func CreateNetworkPartition(name string, params NetworkPartitionParameters, provider *aws.Provider, vpc *ec2.Vpc, defaultRoute *ec2.RouteTable) *pgocomp.Component[*NetworkPartitionComponent] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*NetworkPartitionComponent, error) {
		var response NetworkPartitionComponent
		var subnets []*ec2.Subnet
		var err = errors.Join(
			CreateSubnetAndAssociateToRoute(name+"-subneta", params.SubnetA, provider, vpc, defaultRoute).GetAndThen(ctx, func(subnet *pgocomp.GetComponentResponse[*ec2.Subnet]) error {
				if params.SubnetA.Active {
					response.SubnetA = subnet
					subnets = append(subnets, subnet.Component)
				}
				return nil
			}),
			CreateSubnetAndAssociateToRoute(name+"-subnetb", params.SubnetB, provider, vpc, defaultRoute).GetAndThen(ctx, func(subnet *pgocomp.GetComponentResponse[*ec2.Subnet]) error {
				if params.SubnetB.Active {
					response.SubnetB = subnet
					subnets = append(subnets, subnet.Component)
				}
				return nil
			}),
			CreateLoadBalancerComponent(name+"-lb", params.LoadBalancer, provider, vpc, subnets).GetAndThen(ctx, func(lb *pgocomp.GetComponentResponse[*LoadBalancerComponent]) error {
				if params.LoadBalancer.Active {
					response.LoadBalancer = lb
				}
				return CreateECSClusterComponents(name+"-cluster", params.ECSClusters, provider, subnets, lb.Component).GetAndThen(ctx, func(clusters *pgocomp.GetComponentResponse[[]*ECSClusterComponent]) error {
					response.ECSClusters = clusters
					return nil
				})
			}),
		)
		return &response, err
	})
}

// CreateLoadBalancerComponent takes some paramenters and creates a new Network Partition
func CreateLoadBalancerComponent(name string, params LoadBalancerParameters, provider *aws.Provider, vpc *ec2.Vpc, subnets []*ec2.Subnet) *pgocomp.Component[*LoadBalancerComponent] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*LoadBalancerComponent](name)
	}
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*LoadBalancerComponent, error) {
		var response LoadBalancerComponent
		var err = errors.Join(
			CreateSecurityGroup(name+"-sg", provider, vpc).GetAndThen(ctx, func(sg *pgocomp.GetComponentResponse[*ec2.SecurityGroup]) error {
				response.SecurityGroup = sg
				return errors.Join(
					CreateLoadBalancerAndAssociateToSubnets(name, params.Type, provider, subnets, sg.Component).GetAndThen(ctx, func(loadBalancer *pgocomp.GetComponentResponse[*lb.LoadBalancer]) error {
						response.LoadBalancer = loadBalancer
						ctx.Export(loadBalancer.Name+"-dns", loadBalancer.Component.DnsName)
						return CreateTargetGroups(name+"-tgt", params.TargetGroups, provider, vpc, loadBalancer.Component).GetAndThen(ctx, func(tgs *pgocomp.GetComponentResponse[map[string]*lb.TargetGroup]) error {
							response.targetGroups = tgs
							return CreateListeners(name+"-lis", params.Listeners, provider, loadBalancer.Component, tgs.Component, sg.Component).GetAndThen(ctx, func(ls *pgocomp.GetComponentResponse[[]*lb.Listener]) error {
								response.listeners = ls
								return nil
							})
						})
					}),
				)
			}),
		)
		return &response, err
	})
}

// CreateECSClusterComponents takes some paramenters and creates a new Network Partition
func CreateECSClusterComponents(name string, params []ECSClusterParameters, provider *aws.Provider, subnets []*ec2.Subnet, lbcomp *LoadBalancerComponent) *pgocomp.Component[[]*ECSClusterComponent] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) ([]*ECSClusterComponent, error) {
		var clusters []*ECSClusterComponent
		for i, clusterParams := range params {
			if !clusterParams.Active {
				continue
			}
			if err := CreateECSClusterComponent(name+"-"+strconv.Itoa(i), clusterParams, provider, subnets, lbcomp).GetAndThen(ctx, func(cluster *pgocomp.GetComponentResponse[*ECSClusterComponent]) error {
				clusters = append(clusters, cluster.Component)
				return nil
			}); err != nil {
				return clusters, err
			}
		}
		return clusters, nil
	})
}

// CreateECSClusterComponent takes some paramenters and creates a new Network Partition
func CreateECSClusterComponent(name string, params ECSClusterParameters, provider *aws.Provider, subnets []*ec2.Subnet, lbcomp *LoadBalancerComponent) *pgocomp.Component[*ECSClusterComponent] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*ECSClusterComponent](name)
	}
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*ECSClusterComponent, error) {
		var cc ECSClusterComponent
		var err = errors.Join(
			CreateECSCluster(name, params, provider).GetAndThen(ctx, func(cluster *pgocomp.GetComponentResponse[*ecs.Cluster]) error {
				if !params.Active {
					return nil
				}
				cc.Cluster = cluster
				return CreateECSServiceComponents(name+"-services", params.Services, provider, cluster.Component, subnets, lbcomp).GetAndThen(ctx, func(svcs *pgocomp.GetComponentResponse[[]*ecs.Service]) error {
					cc.Services = svcs
					return nil
				})
			}),
		)
		return &cc, err
	})
}

// CreateECSServiceComponents takes some paramenters and creates a new Network Partition
func CreateECSServiceComponents(name string, params []ECSServiceParameters, provider *aws.Provider, cluster *ecs.Cluster, subnets []*ec2.Subnet, lbcomp *LoadBalancerComponent) *pgocomp.Component[[]*ecs.Service] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) ([]*ecs.Service, error) {
		var response []*ecs.Service
		for i, svcParams := range params {
			if !svcParams.Active {
				continue
			}
			if err := CreateServiceComponent(name+"-"+strconv.Itoa(i), svcParams, provider, cluster, subnets, lbcomp).GetAndThen(ctx, func(svc *pgocomp.GetComponentResponse[*ecs.Service]) error {
				response = append(response, svc.Component)
				return nil
			}); err != nil {
				return nil, err
			}
		}
		return response, nil
	})
}

// CreateServiceComponent creates bla bla bla
func CreateServiceComponent(name string, params ECSServiceParameters, provider *aws.Provider, cluster *ecs.Cluster, subnets []*ec2.Subnet, lbcomp *LoadBalancerComponent) *pgocomp.Component[*ecs.Service] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*ecs.Service](name)
	}
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*ecs.Service, error) {
		var svc *ecs.Service
		dependsOn := []pulumi.Resource{cluster, lbcomp.LoadBalancer.Component}
		for _, subnet := range subnets {
			dependsOn = append(dependsOn, subnet)
		}
		var ecsLaunchType ecs.ServiceLaunchType = ecs.ServiceLaunchTypeFargate
		if params.LaunchType == EC2LaunchType {
			ecsLaunchType = ecs.ServiceLaunchTypeEc2
		}
		var loadBalancerArray ecs.ServiceLoadBalancerArray
		var containerDefinitions ecs.TaskDefinitionContainerDefinitionArray
		for _, containerParam := range params.Containers {
			containerParam.Definition.Name = pulumi.String(containerParam.Name) //overwrides the definition.Name attribute
			containerDefinitions = append(containerDefinitions, containerParam.Definition)
			for _, loadBalancerInfo := range containerParam.LoadBalancerInfo {
				loadBalancerArray = append(loadBalancerArray, ecs.ServiceLoadBalancerArgs{
					ContainerName:  pulumi.String(containerParam.Name),
					ContainerPort:  pulumi.Int(loadBalancerInfo.ContainerPort),
					TargetGroupArn: lbcomp.targetGroups.Component[loadBalancerInfo.TargetGroupLookupName].ID(),
				})
				dependsOn = append(dependsOn, lbcomp.targetGroups.Component[loadBalancerInfo.TargetGroupLookupName])
			}
		}
		var err = CreateFargateTaskDefinition(name+"-task", params, provider, cluster, &containerDefinitions).GetAndThen(ctx, func(taskDef *pgocomp.GetComponentResponse[*ecs.TaskDefinition]) error {
			var customResources []*pulumi.CustomResourceState
			for _, subnet := range subnets {
				customResources = append(customResources, &subnet.CustomResourceState)
			}
			awsc.NewService(name, &ecs.ServiceArgs{
				Cluster:        cluster.ID(),
				DesiredCount:   pulumi.Int(params.DesiredCount),
				LaunchType:     ecsLaunchType,
				TaskDefinition: taskDef.Component.ID(),
				NetworkConfiguration: ecs.ServiceNetworkConfigurationArgs{
					AwsvpcConfiguration: ecs.ServiceAwsVpcConfigurationArgs{
						AssignPublicIp: params.GetPublicIP(),
						Subnets:        awsc.ToIDStringArray(customResources...),
						SecurityGroups: awsc.ToIDStringArray(&lbcomp.SecurityGroup.Component.CustomResourceState),
					},
				},
				LoadBalancers: loadBalancerArray,
				//TODO: fix delete service automatically by forcing it
			},
				pulumi.Provider(provider),
				pulumi.DependsOn(dependsOn)).GetAndThen(ctx, func(service *pgocomp.GetComponentResponse[*ecs.Service]) error {
				svc = service.Component
				return nil
			})
			return nil
		})
		return svc, err
	})
}

// CreateFargateTaskDefinition creates TaskDefinition for the Service
func CreateFargateTaskDefinition(name string, params ECSServiceParameters, provider *aws.Provider, cluster *ecs.Cluster, containerDefinitions *ecs.TaskDefinitionContainerDefinitionArray) *pgocomp.Component[*ecs.TaskDefinition] {

	if params.LaunchType == FargateSpotLaunchType {
		params.LaunchType = FargateLaunchType
	}

	return awsc.NewTaskDefinition(name, &ecs.TaskDefinitionArgs{
		Cpu:    pulumi.String(params.CPU),
		Memory: pulumi.String(params.Memory),
		RequiresCompatibilities: pulumi.StringArray{
			pulumi.String(params.LaunchType),
		},
		ContainerDefinitions: containerDefinitions,
		NetworkMode:          pulumi.String("awsvpc"),
	}, pulumi.Provider(provider), pulumi.ReplaceOnChanges([]string{"*"}), pulumi.DependsOn([]pulumi.Resource{cluster}))
}

// CreateECSCluster creates a new ECSCluster
func CreateECSCluster(name string, params ECSClusterParameters, provider *aws.Provider) *pgocomp.Component[*ecs.Cluster] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*ecs.Cluster](name)
	}
	return awsc.NewCluster(name, &ecs.ClusterArgs{}, pulumi.Provider(provider))
}

// CreateProvider takes a name and a region and returns an aws.Provider Component
func CreateProvider(name string, region string) *pgocomp.Component[*aws.Provider] {
	return awsc.NewProvider(name, &aws.ProviderArgs{Region: pulumi.String(region)})
}

// CreateVPC takes a name, a list of parameters and a provides and returns a Vpc Component
func CreateVPC(name string, params RegionParameters, provider *aws.Provider) *pgocomp.Component[*ec2.Vpc] {
	return awsc.NewVpc(
		name,
		&ec2.VpcArgs{
			CidrBlock:        pulumi.String(params.CidrBlock),
			EnableDnsSupport: pulumi.Bool(true),
		},
		pulumi.Provider(provider).(pulumi.ResourceOption),
	)
}

// CreateListeners creates a new Listener Component
func CreateListeners(name string, params []LBListenerParameters, provider *aws.Provider, loadBalancer *lb.LoadBalancer, tgs map[string]*lb.TargetGroup, sg *ec2.SecurityGroup) *pgocomp.Component[[]*lb.Listener] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) ([]*lb.Listener, error) {
		var response []*lb.Listener
		for i, ps := range params {
			if err := CreateListener(name+"-"+strconv.Itoa(i), ps, provider, loadBalancer, tgs, sg).GetAndThen(ctx, func(l *pgocomp.GetComponentResponse[*lb.Listener]) error {
				if l == nil {
					return nil
				}
				response = append(response, l.Component)
				//Create a Security Group rule that opens the port
				return CreateAndAttachTCPIngressSecurityGroupRule(l.Name+"-sg-rule", provider, sg, ps.Port, ps.Port, []string{"0.0.0.0/0"}).Apply(ctx)
			}); err != nil {
				return response, err
			}
		}
		return response, nil
	})
}

// CreateListener creates a new Listener Component
func CreateListener(name string, params LBListenerParameters, provider *aws.Provider, loadBalancer *lb.LoadBalancer, tgs map[string]*lb.TargetGroup, sg *ec2.SecurityGroup) *pgocomp.Component[*lb.Listener] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*lb.Listener](name)
	}
	return awsc.NewListener(name, &lb.ListenerArgs{
		Port:            pulumi.Int(params.Port),
		LoadBalancerArn: loadBalancer.ID(),
		Protocol:        pulumi.String("HTTP"), //TODO: Listener protocol should match targetGroup. now it only works for HTTP
		DefaultActions: lb.ListenerDefaultActionArray{
			lb.ListenerDefaultActionArgs{
				TargetGroupArn: tgs[params.TargetGroupLookupName].ID(),
				Type:           pulumi.String("forward"),
			},
		},
	}, pulumi.Provider(provider))
}

// CreateTargetGroups creates a new LoadBalancer Component
func CreateTargetGroups(name string, params []LBTargetGroupParameters, provider *aws.Provider, vpc *ec2.Vpc, loadBalancer *lb.LoadBalancer) *pgocomp.Component[map[string]*lb.TargetGroup] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (map[string]*lb.TargetGroup, error) {
		var response = make(map[string]*lb.TargetGroup)
		for i, ps := range params {
			if err := CreateTargetGroup(name+"-"+strconv.Itoa(i), ps, provider, vpc, loadBalancer).GetAndThen(ctx, func(tg *pgocomp.GetComponentResponse[*lb.TargetGroup]) error {
				if tg == nil {
					return nil
				}
				response[ps.LookupName] = tg.Component
				return nil
			}); err != nil {
				return response, err
			}
		}
		return response, nil
	})
}

// CreateTargetGroup creates a new LoadBalancer Component
func CreateTargetGroup(name string, params LBTargetGroupParameters, provider *aws.Provider, vpc *ec2.Vpc, loadBalancer *lb.LoadBalancer) *pgocomp.Component[*lb.TargetGroup] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*lb.TargetGroup](name)
	}
	return awsc.NewTargetGroup(name, &lb.TargetGroupArgs{
		Port:       pulumi.Int(params.Port),
		VpcId:      vpc.ID(),
		Protocol:   pulumi.String("HTTP"),
		TargetType: pulumi.String("ip"),
	}, pulumi.Provider(provider))
}

// CreateLoadBalancerAndAssociateToSubnets creates a new LoadBalancer Component
func CreateLoadBalancerAndAssociateToSubnets(name string, lbType LBType, provider *aws.Provider, subnets []*ec2.Subnet, sg *ec2.SecurityGroup) *pgocomp.Component[*lb.LoadBalancer] {
	var dependsOn []pulumi.Resource
	for _, subnet := range subnets {
		subnet.ID()
		dependsOn = append(dependsOn, subnet)
	}
	var customResources []*pulumi.CustomResourceState
	for _, subnet := range subnets {
		customResources = append(customResources, &subnet.CustomResourceState)
	}
	return awsc.NewLoadBalancer(name, &lb.LoadBalancerArgs{
		LoadBalancerType: pulumi.String(lbType),
		SecurityGroups:   awsc.ToIDStringArray(&sg.CustomResourceState),
		Subnets:          awsc.ToIDStringArray(customResources...),
	}, pulumi.Provider(provider), pulumi.DependsOn(dependsOn))
}

// CreateSecurityGroup takes a name and a vpc and returns a SecurityGroup Component
func CreateSecurityGroup(name string, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.Component[*ec2.SecurityGroup] {
	return awsc.NewSecurityGroup(
		name,
		&ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateAndAttachTCPIngressSecurityGroupRule takes a name, a security group, some nework parameters and returns a SecurityGroupRule Component
func CreateAndAttachTCPIngressSecurityGroupRule(name string, provider *aws.Provider, sg *ec2.SecurityGroup, fromPort, toPort int, cidrBlocks []string) *pgocomp.Component[*ec2.SecurityGroupRule] {
	return awsc.NewSecurityGroupRule(
		name, &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			Protocol:        pulumi.String(("tcp")),
			SecurityGroupId: sg.ID(),
			CidrBlocks:      pulumi.ToStringArray(cidrBlocks),
			FromPort:        pulumi.Int(fromPort),
			ToPort:          pulumi.Int(toPort),
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{sg}))
}

// CreateSubnetAndAssociateToRoute takes a name, some parameters and a vpc and returns a SubnetComponent
func CreateSubnetAndAssociateToRoute(name string, params SubnetParameters, provider *aws.Provider, vpc *ec2.Vpc, rt *ec2.RouteTable) *pgocomp.Component[*ec2.Subnet] {
	return pgocomp.NewComponent(name, func(ctx *pulumi.Context) (*ec2.Subnet, error) {
		var subnet *ec2.Subnet
		err := CreateSubnet(name, params, provider, vpc, rt).GetAndThen(ctx, func(s *pgocomp.GetComponentResponse[*ec2.Subnet]) error {
			if rt == nil {
				return nil
			}
			subnet = s.Component
			return AssociateRouteTableToSubnet(name+"-route", provider, subnet, rt).Apply(ctx)
		})

		return subnet, err
	})
}

// CreateSubnet takes a name, some parameters and a vpc and returns a SubnetComponent
func CreateSubnet(name string, params SubnetParameters, provider *aws.Provider, vpc *ec2.Vpc, rt *ec2.RouteTable) *pgocomp.Component[*ec2.Subnet] {
	if !params.Active {
		return pgocomp.NewInactiveComponent[*ec2.Subnet](name)
	}
	dependsOn := []pulumi.Resource{vpc}
	if rt != nil {
		dependsOn = append(dependsOn, rt)
	}
	return awsc.NewSubnet(
		name,
		&ec2.SubnetArgs{
			AvailabilityZone: pulumi.String(params.AvailabilityZone),
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(params.CidrBlock),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn(dependsOn),
	)
}

// CreateInternetGateway takes a name, a vpc and returns an InternetGateway component
func CreateInternetGateway(name string, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.Component[*ec2.InternetGateway] {
	return awsc.NewInternetGateway(
		name,
		&ec2.InternetGatewayArgs{},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// AttachInternetGatewayToVPC takes a name, a vpc, an internet gateway and returnas an InternetGatewayAttachment Component
func AttachInternetGatewayToVPC(name string, provider *aws.Provider, vpc *ec2.Vpc, igw *ec2.InternetGateway) *pgocomp.Component[*ec2.InternetGatewayAttachment] {
	return awsc.NewInternetGatewayAttachment(
		name,
		&ec2.InternetGatewayAttachmentArgs{
			VpcId:             vpc.ID(),
			InternetGatewayId: igw.ID(),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{vpc, igw}),
	)
}

// CreateRouteTable takes a name and a vpc and returns a RouteTable Component
func CreateRouteTable(name string, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.Component[*ec2.RouteTable] {
	return awsc.NewRouteTable(
		name,
		&ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateAndAttachDefaultRoute takes a name, a route table, an internetgateway and an internet gateway attachment and returns a Route Component
func CreateAndAttachDefaultRoute(name string, provider *aws.Provider, rt *ec2.RouteTable, igw *ec2.InternetGateway, iga *ec2.InternetGatewayAttachment) *pgocomp.Component[*ec2.Route] {
	return awsc.NewRoute(
		name,
		&ec2.RouteArgs{
			RouteTableId:         rt.ID(),
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			GatewayId:            igw.ID(),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{igw, iga}),
	)
}

// AssociateRouteTableToSubnet associates a subnet to a route table
func AssociateRouteTableToSubnet(name string, provider *aws.Provider, subnet *ec2.Subnet, routeTable *ec2.RouteTable) *pgocomp.Component[*ec2.RouteTableAssociation] {
	return awsc.NewRouteTableAssociation(name, &ec2.RouteTableAssociationArgs{
		RouteTableId: routeTable.ID(),
		SubnetId:     subnet.ID(),
	}, pulumi.Provider(provider))
}
