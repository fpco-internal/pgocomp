package awscinfra

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/fpco-internal/pgocomp/pkg/awsc"

	"github.com/fpco-internal/pgocomp"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lb"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// New create a new infrastructure
func New(params InfraParameters) *pgocomp.ComponentWithMeta[*InfraComponent] {
	return pgocomp.NewComponentWithMeta(params.Meta, func(ctx *pulumi.Context, name string) (response *InfraComponent, err error) {
		response = &InfraComponent{
			Vpcs: make(map[string]*pgocomp.GetComponentWithMetaResponse[*VpcComponent]),
		}
		for _, vpcParams := range params.Vpcs {
			err = CreateVpcComponent(vpcParams).GetAndThen(ctx, func(vpc *pgocomp.GetComponentWithMetaResponse[*VpcComponent]) error {
				response.Vpcs[vpcParams.Name] = vpc
				return nil
			})
			if err != nil {
				return
			}
		}
		return
	})
}

// CreateVpcComponent creates a Region Component thar comprises of a Vpc and its subcomponents
func CreateVpcComponent(params VpcParameters) *pgocomp.ComponentWithMeta[*VpcComponent] {
	return pgocomp.NewComponentWithMeta(params.Meta, func(ctx *pulumi.Context, name string) (response *VpcComponent, err error) {
		response = &VpcComponent{
			Partitions:   make(map[string]*pgocomp.GetComponentWithMetaResponse[*NetworkPartitionComponent]),
			Certificates: make(map[string]*pgocomp.GetComponentWithMetaResponse[*acm.Certificate]),
		}
		err = errors.Join(
			CreateProvider(
				params.Provider.Meta,
				params.Provider.Region).GetAndThen(ctx, func(provider *pgocomp.GetComponentWithMetaResponse[*aws.Provider]) error {
				response.Provider = provider
				return CreateVPC(params.Meta, params, provider.Component).GetAndThen(ctx, func(vpc *pgocomp.GetComponentWithMetaResponse[*ec2.Vpc]) error {
					response.Vpc = vpc
					return errors.Join(
						CreateInternetGateway(pgocomp.Meta{
							Name: vpc.Meta.Name + "-igw",
						}, provider.Component, vpc.Component).GetAndThen(ctx, func(igw *pgocomp.GetComponentWithMetaResponse[*ec2.InternetGateway]) error {
							response.Gateway.InternetGateway = igw
							return AttachInternetGatewayToVPC(pgocomp.Meta{
								Name: vpc.Meta.Name + "-igw-attach",
							}, provider.Component, vpc.Component, igw.Component).GetAndThen(ctx, func(iga *pgocomp.GetComponentWithMetaResponse[*ec2.InternetGatewayAttachment]) error {
								response.Gateway.VpcGatewayAttachment = iga
								return CreateRouteTable(pgocomp.Meta{
									Name: vpc.Meta.Name + "-igw-routetable",
								}, provider.Component, vpc.Component).GetAndThen(ctx, func(rt *pgocomp.GetComponentWithMetaResponse[*ec2.RouteTable]) error {
									response.Gateway.RouteTable = rt
									return errors.Join(
										CreateAndAttachDefaultRoute(pgocomp.Meta{
											Name: vpc.Meta.Name + "-igw-default-route",
										}, provider.Component, rt.Component, igw.Component, iga.Component).GetAndThen(ctx, func(r *pgocomp.GetComponentWithMetaResponse[*ec2.Route]) error {
											response.Gateway.DefaultRoute = r
											return nil
										}),
									)
								})
							})
						}),
						func() (err error) {
							for _, certificate := range params.Certificates {
								err = CreateCertificate(certificate.Meta, certificate, provider.Component).
									GetAndThen(ctx, func(cert *pgocomp.GetComponentWithMetaResponse[*acm.Certificate]) (err error) {
										response.Certificates[cert.Meta.Name] = cert
										return
									})
								if err != nil {
									break
								}
							}
							return
						}(),
						func() (err error) {
							for _, partition := range params.Partitions {
								certs := make(map[string]*acm.Certificate)
								for k, v := range response.Certificates {
									certs[k] = v.Component
								}
								err = CreateNetworkPartition(
									partition.Meta, partition, provider.Component, vpc.Component, response.Gateway.RouteTable.Component, certs).
									GetAndThen(ctx, func(npc *pgocomp.GetComponentWithMetaResponse[*NetworkPartitionComponent]) error {
										response.Partitions[npc.Meta.Name] = npc
										return nil
									})
								if err != nil {
									break
								}
							}
							return
						}(),
					)
				})
			}),
		)
		return
	},
	)
}

// CreateNetworkPartition takes some paramenters and creates a new Network Partition
func CreateNetworkPartition(meta pgocomp.Meta, params NetworkPartitionParameters, provider *aws.Provider, vpc *ec2.Vpc, rt *ec2.RouteTable, certs map[string]*acm.Certificate) *pgocomp.ComponentWithMeta[*NetworkPartitionComponent] {
	return pgocomp.NewComponentWithMeta(meta, func(ctx *pulumi.Context, name string) (response *NetworkPartitionComponent, err error) {
		response = &NetworkPartitionComponent{
			Subnets:       make(map[string]*pgocomp.GetComponentWithMetaResponse[*ec2.Subnet]),
			LoadBalancers: make(map[string]*pgocomp.GetComponentWithMetaResponse[*LoadBalancerComponent]),
			TargetGroups:  make(map[string]*pgocomp.GetComponentWithMetaResponse[*lb.TargetGroup]),
			ECSClusters:   make(map[string]*pgocomp.GetComponentWithMetaResponse[*ECSClusterComponent]),
		}
		azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
			State: pulumi.StringRef("available"),
		}, pulumi.Provider(provider))
		if err != nil {
			return nil, err
		}
		err = errors.Join(
			//CreateSubnets
			func() (err error) {
				for i, subnet := range params.Subnets {
					var srt = rt
					if !params.IsPublic {
						srt = nil
					}
					err = CreateSubnetAndAssociateToRoute(subnet.Meta, subnet, azs.Names[i%len(azs.Names)], provider, vpc, srt).GetAndThen(ctx, func(subnet *pgocomp.GetComponentWithMetaResponse[*ec2.Subnet]) error {
						response.Subnets[subnet.Meta.Name] = subnet
						return nil
					})
					if err != nil {
						return
					}
				}
				return
			}(),

			//CreateTargetGroups
			func() (err error) {
				for _, tg := range params.LBTargetGroups {
					err = CreateTargetGroup(tg.Meta, tg, provider, vpc).GetAndThen(ctx, func(tgc *pgocomp.GetComponentWithMetaResponse[*lb.TargetGroup]) error {
						response.TargetGroups[tg.Meta.Name] = tgc
						return nil
					})
					if err != nil {
						return
					}
				}
				return
			}(),

			//CreateLoadBalancers
			func() (err error) {
				var subnets []*ec2.Subnet
				for _, s := range response.Subnets {
					subnets = append(subnets, s.Component)
				}
				for _, loadBalancer := range params.LoadBalancers {
					err = CreateLoadBalancerComponent(loadBalancer.Meta, loadBalancer, provider, vpc, subnets, response.TargetGroups, certs).GetAndThen(ctx, func(lbc *pgocomp.GetComponentWithMetaResponse[*LoadBalancerComponent]) error {
						response.LoadBalancers[loadBalancer.Meta.Name] = lbc
						return nil
					})
					if err != nil {
						return
					}
				}
				return
			}(),

			//CreateClusters
			func() (err error) {

				//Collect subnets
				var subnets []*ec2.Subnet
				for _, s := range response.Subnets {
					subnets = append(subnets, s.Component)
				}

				//Collect target groups
				var tgs = make(map[string]*lb.TargetGroup)
				for k, v := range response.TargetGroups {
					tgs[k] = v.Component
				}

				//Collect security groups
				var sgs []*ec2.SecurityGroup
				for _, b := range response.LoadBalancers {
					sgs = append(sgs, b.Component.SecurityGroup.Component)
				}

				for _, cluster := range params.ECSClusters {
					err = CreateECSClusterComponent(cluster.Meta, cluster, provider, vpc, subnets, tgs, sgs).GetAndThen(ctx, func(cls *pgocomp.GetComponentWithMetaResponse[*ECSClusterComponent]) error {
						response.ECSClusters[cls.Meta.Name] = cls
						return nil
					})
					if err != nil {
						return
					}
				}
				return
			}(),
		)
		return
	})
}

// CreateLoadBalancerComponent takes some paramenters and creates a new Network Partition
func CreateLoadBalancerComponent(meta pgocomp.Meta, params LoadBalancerParameters, provider *aws.Provider, vpc *ec2.Vpc, subnets []*ec2.Subnet, tgs map[string]*pgocomp.GetComponentWithMetaResponse[*lb.TargetGroup], certs map[string]*acm.Certificate) *pgocomp.ComponentWithMeta[*LoadBalancerComponent] {
	return pgocomp.NewComponentWithMeta(meta, func(ctx *pulumi.Context, name string) (*LoadBalancerComponent, error) {
		var response LoadBalancerComponent = LoadBalancerComponent{
			Listeners: make(map[string]*pgocomp.GetComponentWithMetaResponse[*lb.Listener]),
		}
		var err = errors.Join(
			CreateSecurityGroup(pgocomp.Meta{
				Name: meta.Name + "-sg",
			}, provider, vpc).GetAndThen(ctx, func(sg *pgocomp.GetComponentWithMetaResponse[*ec2.SecurityGroup]) error {
				response.SecurityGroup = sg
				return errors.Join(
					CreateLoadBalancerAndAssociateToSubnets(meta, params.Type, provider, subnets, sg.Component).GetAndThen(ctx, func(loadBalancer *pgocomp.GetComponentWithMetaResponse[*lb.LoadBalancer]) error {
						response.LoadBalancer = loadBalancer
						ctx.Export(loadBalancer.Meta.FullName()+"-dns", loadBalancer.Component.DnsName)
						for _, lis := range params.Listeners {
							tg, ok := tgs[lis.TargetGroupLookupName]
							if !ok {
								return fmt.Errorf("Target group Lookup Name %s not found", lis.TargetGroupLookupName)
							}
							if err := CreateListener(
								lis.Meta, lis, provider, loadBalancer.Component, tg.Component, sg.Component, certs).GetAndThen(ctx, func(l *pgocomp.GetComponentWithMetaResponse[*lb.Listener]) error {
								response.Listeners[l.Meta.Name] = l
								return CreateAndAttachTCPIngressSecurityGroupRule(
									pgocomp.Meta{Name: lis.Meta.Name + "-rule"},
									provider, sg.Component, lis.Port, lis.Port, []string{"0.0.0.0/0"}).Apply(ctx)
							}); err != nil {
								return err
							}
						}
						return nil
					}),
				)
			}),
		)
		return &response, err
	})
}

// CreateECSClusterComponent takes some paramenters and creates a new Network Partition
func CreateECSClusterComponent(meta pgocomp.Meta, params ECSClusterParameters, provider *aws.Provider, vpc *ec2.Vpc, subnets []*ec2.Subnet, tgs map[string]*lb.TargetGroup, sgs []*ec2.SecurityGroup) *pgocomp.ComponentWithMeta[*ECSClusterComponent] {
	return pgocomp.NewComponentWithMeta(meta, func(ctx *pulumi.Context, name string) (response *ECSClusterComponent, err error) {
		response = &ECSClusterComponent{
			FargateServices: make(map[string]*pgocomp.GetComponentWithMetaResponse[*ecs.Service]),
		}
		err = errors.Join(
			CreateECSCluster(
				meta, params, provider).GetAndThen(ctx, func(cluster *pgocomp.GetComponentWithMetaResponse[*ecs.Cluster]) error {
				response.Cluster = cluster
				for _, svcParams := range params.Services {
					err := CreateEcsFargateServiceComponent(
						svcParams.Meta,
						svcParams,
						provider,
						vpc,
						cluster.Component,
						subnets,
						tgs,
					).GetAndThen(ctx, func(svc *pgocomp.GetComponentWithMetaResponse[*ecs.Service]) error {
						response.FargateServices[svc.Meta.Name] = svc
						return nil
					})
					if err != nil {
						return err
					}
				}
				return nil
			}),
		)
		return
	})
}

func matchOrPanic(pattern string, str string) bool {
	b, err := regexp.MatchString(pattern, str)
	if err != nil {
		panic(err)
	}
	return b
}

// CreateCertificate creates a new certificate with a CertificateParameters
func CreateCertificate(meta pgocomp.Meta, params CertificateParameters, provider *aws.Provider) *pgocomp.ComponentWithMeta[*acm.Certificate] {
	return awsc.NewCertificate(meta, &acm.CertificateArgs{
		DomainName:       pulumi.String(params.Domain),
		ValidationMethod: pulumi.String(params.ValidationMethod),
	}, pulumi.Provider(provider), pulumi.Protect(meta.Protect))
}

// CreateEcsFargateServiceComponent is
func CreateEcsFargateServiceComponent(
	meta pgocomp.Meta,
	params ECSServiceParameters,
	provider *aws.Provider,
	vpc *ec2.Vpc,
	cluster *ecs.Cluster,
	subnets []*ec2.Subnet,
	targetGroups map[string]*lb.TargetGroup,
) *pgocomp.ComponentWithMeta[*ecs.Service] {
	return pgocomp.NewComponentWithMeta(meta, func(ctx *pulumi.Context, name string) (response *ecs.Service, err error) {

		//Security group for the Service
		err = CreateSecurityGroup(pgocomp.Meta{Name: meta.Name + "-sg"}, provider, vpc).GetAndThen(ctx, func(sg *pgocomp.GetComponentWithMetaResponse[*ec2.SecurityGroup]) (err error) {
			var containerDefinitions string
			containerDefinitions, err = params.ECSTaskDefinitionContainerDefinitionArray()
			if err != nil {
				return
			}
			err = awsc.NewEcsTaskDefinition(
				pgocomp.Meta{
					Name: meta.Name + "-task",
				},
				&ecs.TaskDefinitionArgs{
					NetworkMode: pulumi.String("awsvpc"),
					RequiresCompatibilities: pulumi.StringArray{
						pulumi.String("FARGATE"),
					},
					Family:               pulumi.String("PULUMI-AUTO"),
					Cpu:                  pulumi.String(strconv.Itoa(params.CPU)),
					Memory:               pulumi.String(strconv.Itoa(params.Memory)),
					ContainerDefinitions: pulumi.String(containerDefinitions),
				}, func() []pulumi.ResourceOption {
					dependsOn := []pulumi.Resource{cluster}
					for _, subnet := range subnets {
						dependsOn = append(dependsOn, subnet)
					}
					for _, c := range params.Containers {
						for _, p := range c.PortMappings {
							if tg, ok := targetGroups[p.TargetGroupLookupName]; !ok {
								dependsOn = append(dependsOn, tg)

							}
						}
					}
					return []pulumi.ResourceOption{
						pulumi.Provider(provider), pulumi.Protect(meta.Protect),
						pulumi.DependsOn(dependsOn),
						//					pulumi.ReplaceOnChanges([]string{"*"}),
					}
				}()...).GetAndThen(ctx, func(taskDef *pgocomp.GetComponentWithMetaResponse[*ecs.TaskDefinition]) error {
				return awsc.NewECSService(params.Meta, &ecs.ServiceArgs{
					Name:           pulumi.String(params.Name),
					Cluster:        cluster.ID(),
					DesiredCount:   pulumi.Int(params.DesiredCount),
					LaunchType:     pulumi.String("FARGATE"),
					TaskDefinition: taskDef.Component.ID(),
					NetworkConfiguration: ecs.ServiceNetworkConfigurationArgs{
						AssignPublicIp: pulumi.Bool(params.AssignPublicIP),
						Subnets: func() (array pulumi.StringArray) {
							for _, subnet := range subnets {
								array = append(array, subnet.ID())
							}
							return
						}(),
						SecurityGroups: pulumi.StringArray{sg.Component.ID()},
					},
					LoadBalancers: func() (array ecs.ServiceLoadBalancerArray) {
						for _, c := range params.Containers {
							for _, p := range c.PortMappings {
								if tg, ok := targetGroups[p.TargetGroupLookupName]; ok {
									if err := CreateSecurityGroupRuleForTargetGroup(
										pgocomp.Meta{Name: meta.Name + "-sg-" + p.TargetGroupLookupName + "-rule"},
										provider,
										sg.Component,
										tg,
										[]string{"0.0.0.0/0"},
									).Apply(ctx); err != nil {
										panic(err)
									}
									array = append(array, ecs.ServiceLoadBalancerArgs{
										ContainerName:  pulumi.String(c.Name),
										ContainerPort:  pulumi.Int(p.ContainerPort),
										TargetGroupArn: tg.ID(),
									})
								}
							}
						}
						return
					}(),
				},
					pulumi.Provider(provider), pulumi.Protect(meta.Protect),
					//				pulumi.ReplaceOnChanges([]string{"*"}),
					pulumi.DependsOn([]pulumi.Resource{cluster, taskDef.Component})).GetAndThen(ctx, func(svc *pgocomp.GetComponentWithMetaResponse[*ecs.Service]) error {
					response = svc.Component
					return nil
				})
			})
			return
		})
		return
	})
}

// CreateECSCluster creates a new ECSCluster
func CreateECSCluster(meta pgocomp.Meta, params ECSClusterParameters, provider *aws.Provider) *pgocomp.ComponentWithMeta[*ecs.Cluster] {
	return awsc.NewCluster(meta, &ecs.ClusterArgs{}, pulumi.Provider(provider), pulumi.Protect(meta.Protect))
}

// CreateProvider takes a name and a region and returns an aws.Provider Component
func CreateProvider(meta pgocomp.Meta, region string) *pgocomp.ComponentWithMeta[*aws.Provider] {
	return awsc.NewProvider(meta, &aws.ProviderArgs{Region: pulumi.String(region)})
}

// CreateVPC takes a meta, a list of parameters and a provides and returns a Vpc Component
func CreateVPC(meta pgocomp.Meta, params VpcParameters, provider *aws.Provider) *pgocomp.ComponentWithMeta[*ec2.Vpc] {
	return awsc.NewVpc(
		meta,
		&ec2.VpcArgs{
			CidrBlock:        pulumi.String(params.CidrBlock),
			EnableDnsSupport: pulumi.Bool(true),
		},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect).(pulumi.ResourceOption),
	)
}

// CreateListener creates a new Listener Component
func CreateListener(meta pgocomp.Meta, params LBListenerParameters, provider *aws.Provider, loadBalancer *lb.LoadBalancer, tg *lb.TargetGroup, sg *ec2.SecurityGroup, certs map[string]*acm.Certificate) *pgocomp.ComponentWithMeta[*lb.Listener] {

	return pgocomp.NewComponentWithMeta[*lb.Listener](meta, func(ctx *pulumi.Context, name string) (response *lb.Listener, err error) {
		args := &lb.ListenerArgs{
			Port:            pulumi.Int(params.Port),
			LoadBalancerArn: loadBalancer.ID(),
			Protocol:        pulumi.String(params.Protocol),
			DefaultActions: lb.ListenerDefaultActionArray{
				lb.ListenerDefaultActionArgs{
					TargetGroupArn: tg.ID(),
					Type:           pulumi.String("forward"),
				},
			},
		}
		if cert, ok := certs[params.CertificateLookupName]; ok {
			args.CertificateArn = cert.ID()
		}
		err = awsc.NewListener(meta, args, pulumi.Provider(provider), pulumi.Protect(meta.Protect)).GetAndThen(ctx, func(l *pgocomp.GetComponentWithMetaResponse[*lb.Listener]) (err error) {
			response = l.Component
			for _, rule := range params.Rules {
				var conditions lb.ListenerRuleConditionArray
				for _, condition := range rule.Conditions {
					switch condition.RuleConditionType {
					case PathPattern:
						conditions = append(conditions, lb.ListenerRuleConditionArgs{
							PathPattern: lb.ListenerRuleConditionPathPatternArgs{
								Values: func() (array pulumi.StringArray) {
									for _, pattern := range condition.PathPatterns {
										array = append(array, pulumi.String(pattern))
									}
									return
								}(),
							},
						})
					case HostHeader:
						conditions = append(conditions, lb.ListenerRuleConditionArgs{
							HostHeader: lb.ListenerRuleConditionHostHeaderArgs{
								Values: func() (array pulumi.StringArray) {
									for _, pattern := range condition.HostHeaders {
										array = append(array, pulumi.String(pattern))
									}
									return
								}(),
							},
						})
					case HTTPHeader:
						conditions = append(conditions, lb.ListenerRuleConditionArgs{
							HttpHeader: lb.ListenerRuleConditionHttpHeaderArgs{
								HttpHeaderName: pulumi.String(condition.HTTPHeader.Name),
								Values: func() (array pulumi.StringArray) {
									for _, pattern := range condition.HTTPHeader.Values {
										array = append(array, pulumi.String(pattern))
									}
									return
								}(),
							},
						})
					case QueryString:
						conditions = append(conditions, lb.ListenerRuleConditionArgs{
							QueryStrings: func() (array lb.ListenerRuleConditionQueryStringArray) {
								for _, pattern := range condition.QueryString {
									array = append(array, lb.ListenerRuleConditionQueryStringArgs{
										Key:   pulumi.String(pattern.Key),
										Value: pulumi.String(pattern.Value),
									})
								}
								return
							}(),
						})
					case SourceIP:
						conditions = append(conditions, lb.ListenerRuleConditionArgs{
							SourceIp: lb.ListenerRuleConditionSourceIpArgs{
								Values: func() (array pulumi.StringArray) {
									for _, pattern := range condition.SourceIPs {
										array = append(array, pulumi.String(pattern))
									}
									return
								}(),
							},
						})
					}
				}
				err = awsc.NewListenerRule(rule.Meta, &lb.ListenerRuleArgs{
					Actions: lb.ListenerRuleActionArray{
						lb.ListenerRuleActionArgs{
							TargetGroupArn: tg.ID(),
							Type:           pulumi.String("forward"),
						},
					},
					ListenerArn: l.Component.ID(),
					Conditions:  conditions,
					Priority:    pulumi.Int(rule.Priority),
				},
					pulumi.Provider(provider), pulumi.Protect(rule.Protect)).
					Apply(ctx)
			}
			return
		})
		return
	})

}

// CreateTargetGroup creates a new LoadBalancer Component
func CreateTargetGroup(meta pgocomp.Meta, params LBTargetGroupParameters, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.ComponentWithMeta[*lb.TargetGroup] {
	return awsc.NewTargetGroup(meta, &lb.TargetGroupArgs{
		Port:       pulumi.Int(params.Port),
		VpcId:      vpc.ID(),
		Protocol:   pulumi.String(params.Protocol),
		TargetType: pulumi.String("ip"),
	}, pulumi.Provider(provider), pulumi.Protect(meta.Protect))
}

// CreateLoadBalancerAndAssociateToSubnets creates a new LoadBalancer Component
func CreateLoadBalancerAndAssociateToSubnets(meta pgocomp.Meta, lbType LBType, provider *aws.Provider, subnets []*ec2.Subnet, sg *ec2.SecurityGroup) *pgocomp.ComponentWithMeta[*lb.LoadBalancer] {
	var dependsOn []pulumi.Resource
	for _, subnet := range subnets {
		subnet.ID()
		dependsOn = append(dependsOn, subnet)
	}
	var customResources []*pulumi.CustomResourceState
	for _, subnet := range subnets {
		customResources = append(customResources, &subnet.CustomResourceState)
	}
	return awsc.NewLoadBalancer(meta, &lb.LoadBalancerArgs{
		LoadBalancerType: pulumi.String(lbType),
		SecurityGroups:   awsc.ToIDStringArray(&sg.CustomResourceState),
		Subnets:          awsc.ToIDStringArray(customResources...),
	}, pulumi.Provider(provider), pulumi.Protect(meta.Protect), pulumi.DependsOn(dependsOn))
}

// CreateSecurityGroup takes a name and a vpc and returns a SecurityGroup Component
func CreateSecurityGroup(meta pgocomp.Meta, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.ComponentWithMeta[*ec2.SecurityGroup] {
	return awsc.NewSecurityGroup(
		meta,
		&ec2.SecurityGroupArgs{
			VpcId: vpc.ID(),
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
					Ipv6CidrBlocks: pulumi.StringArray{
						pulumi.String("::/0"),
					},
				},
			},
		},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect),
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateAndAttachTCPIngressSecurityGroupRule takes a meta, a security group, some nework parameters and returns a SecurityGroupRule Component
func CreateAndAttachTCPIngressSecurityGroupRule(meta pgocomp.Meta, provider *aws.Provider, sg *ec2.SecurityGroup, fromPort, toPort int, cidrBlocks []string) *pgocomp.ComponentWithMeta[*ec2.SecurityGroupRule] {
	return awsc.NewSecurityGroupRule(
		meta, &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			Protocol:        pulumi.String(("tcp")),
			SecurityGroupId: sg.ID(),
			CidrBlocks:      pulumi.ToStringArray(cidrBlocks),
			FromPort:        pulumi.Int(fromPort),
			ToPort:          pulumi.Int(toPort),
		}, pulumi.Provider(provider), pulumi.Protect(meta.Protect), pulumi.DependsOn([]pulumi.Resource{sg}))
}

// CreateSecurityGroupRuleForTargetGroup ...
func CreateSecurityGroupRuleForTargetGroup(meta pgocomp.Meta, provider *aws.Provider, sg *ec2.SecurityGroup, tg *lb.TargetGroup, cidrBlocks []string) *pgocomp.ComponentWithMeta[*ec2.SecurityGroupRule] {
	return awsc.NewSecurityGroupRule(
		meta, &ec2.SecurityGroupRuleArgs{
			Type:            pulumi.String("ingress"),
			Protocol:        pulumi.String(("tcp")),
			SecurityGroupId: sg.ID(),
			CidrBlocks:      pulumi.ToStringArray(cidrBlocks),
			FromPort:        tg.Port.Elem(),
			ToPort:          tg.Port.Elem(),
		}, pulumi.Provider(provider), pulumi.Protect(meta.Protect), pulumi.DependsOn([]pulumi.Resource{sg}))
}

// CreateSubnetAndAssociateToRoute takes a meta, some parameters and a vpc and returns a SubnetComponent
func CreateSubnetAndAssociateToRoute(meta pgocomp.Meta, params SubnetParameters, az string, provider *aws.Provider, vpc *ec2.Vpc, rt *ec2.RouteTable) *pgocomp.ComponentWithMeta[*ec2.Subnet] {
	return pgocomp.NewComponentWithMeta(meta, func(ctx *pulumi.Context, name string) (*ec2.Subnet, error) {
		var subnet *ec2.Subnet
		err := CreateSubnet(meta, params, az, provider, vpc, rt).GetAndThen(ctx, func(s *pgocomp.GetComponentWithMetaResponse[*ec2.Subnet]) error {
			subnet = s.Component
			return AssociateRouteTableToSubnet(pgocomp.Meta{Name: meta.Name + "-route-association"}, provider, subnet, rt).Apply(ctx)
		})
		return subnet, err
	})
}

// CreateSubnet takes a meta, some parameters and a vpc and returns a SubnetComponent
func CreateSubnet(meta pgocomp.Meta, params SubnetParameters, az string, provider *aws.Provider, vpc *ec2.Vpc, rt *ec2.RouteTable) *pgocomp.ComponentWithMeta[*ec2.Subnet] {
	dependsOn := []pulumi.Resource{vpc}
	if rt != nil {
		dependsOn = append(dependsOn, rt)
	}
	return awsc.NewSubnet(
		meta,
		&ec2.SubnetArgs{
			AvailabilityZone: pulumi.String(az),
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(params.CidrBlock),
		},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect),
		pulumi.DependsOn(dependsOn),
	)
}

// CreateInternetGateway takes a meta, a vpc and returns an InternetGateway component
func CreateInternetGateway(meta pgocomp.Meta, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.ComponentWithMeta[*ec2.InternetGateway] {
	return awsc.NewInternetGateway(
		meta,
		&ec2.InternetGatewayArgs{},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect),
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// AttachInternetGatewayToVPC takes a meta, a vpc, an internet gateway and returnas an InternetGatewayAttachment Component
func AttachInternetGatewayToVPC(meta pgocomp.Meta, provider *aws.Provider, vpc *ec2.Vpc, igw *ec2.InternetGateway) *pgocomp.ComponentWithMeta[*ec2.InternetGatewayAttachment] {
	return awsc.NewInternetGatewayAttachment(
		meta,
		&ec2.InternetGatewayAttachmentArgs{
			VpcId:             vpc.ID(),
			InternetGatewayId: igw.ID(),
		},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect),
		pulumi.DependsOn([]pulumi.Resource{vpc, igw}),
	)
}

// CreateRouteTable takes a name and a vpc and returns a RouteTable Component
func CreateRouteTable(meta pgocomp.Meta, provider *aws.Provider, vpc *ec2.Vpc) *pgocomp.ComponentWithMeta[*ec2.RouteTable] {
	return awsc.NewRouteTable(
		meta,
		&ec2.RouteTableArgs{
			VpcId: vpc.ID(),
		},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect),
		pulumi.DependsOn([]pulumi.Resource{vpc}),
	)
}

// CreateAndAttachDefaultRoute takes a meta, a route table, an internetgateway and an internet gateway attachment and returns a Route Component
func CreateAndAttachDefaultRoute(meta pgocomp.Meta, provider *aws.Provider, rt *ec2.RouteTable, igw *ec2.InternetGateway, iga *ec2.InternetGatewayAttachment) *pgocomp.ComponentWithMeta[*ec2.Route] {
	return awsc.NewRoute(
		meta,
		&ec2.RouteArgs{
			RouteTableId:         rt.ID(),
			DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
			GatewayId:            igw.ID(),
		},
		pulumi.Provider(provider), pulumi.Protect(meta.Protect),
		pulumi.DependsOn([]pulumi.Resource{igw, iga}),
	)
}

// AssociateRouteTableToSubnet associates a subnet to a route table
func AssociateRouteTableToSubnet(meta pgocomp.Meta, provider *aws.Provider, subnet *ec2.Subnet, routeTable *ec2.RouteTable) *pgocomp.ComponentWithMeta[*ec2.RouteTableAssociation] {
	return awsc.NewRouteTableAssociation(meta, &ec2.RouteTableAssociationArgs{
		RouteTableId: routeTable.ID(),
		SubnetId:     subnet.ID(),
	}, pulumi.Provider(provider), pulumi.Protect(meta.Protect))
}
