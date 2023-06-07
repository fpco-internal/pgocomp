package awscinfra

import (
	"log"
	"sync"
	"testing"

	"github.com/fpco-internal/pgocomp"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

type mocks int

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func checkTags(t *testing.T, urn pulumi.URNOutput, tags pulumi.StringMapOutput, checks ...func(t *testing.T, urn pulumi.URN, tags map[string]string)) {
	pulumi.All(urn, tags).ApplyT(func(all []interface{}) error {
		urn := all[0].(pulumi.URN)
		tags := all[1].(map[string]string)
		for _, check := range checks {
			check(t, urn, tags)
		}
		return nil
	})
}

func AllTagNamesExist(tagNames ...string) func(t *testing.T, urn pulumi.URN, tags map[string]string) {
	return func(t *testing.T, urn pulumi.URN, tags map[string]string) {
		for _, tagname := range tagNames {
			assert.Containsf(t, tags, tagname, "missing %v tag on component %v", tagname, urn)
		}
	}
}

func TestNewBasicNetworkComponent(t *testing.T) {
	type args struct {
		name   string
		params RegionParameters
	}
	tests := []struct {
		name      string
		args      args
		test      func(*testing.T, *pgocomp.GetComponentResponse[*RegionComponent])
		wantError bool
	}{
		{
			name: "create a basic network",
			args: args{},
			test: func(t *testing.T, bn *pgocomp.GetComponentResponse[*RegionComponent]) {

				log.Printf("Testing")

				var wg sync.WaitGroup
				runTest := func(runTest func()) {
					wg.Add(1)
					defer wg.Done()
					runTest()
				}
				runTest(func() {
					var tagNamesCheck = AllTagNamesExist()

					checkTags(t, bn.Component.Vpc.Component.URN(), bn.Component.Vpc.Component.Tags, tagNamesCheck)
					checkTags(t, bn.Component.Partitions.Public.Component.SubnetA.Component.URN(), bn.Component.Partitions.Public.Component.SubnetA.Component.Tags, tagNamesCheck)
					checkTags(t, bn.Component.Partitions.Private.Component.SubnetA.Component.URN(), bn.Component.Partitions.Private.Component.SubnetA.Component.Tags, tagNamesCheck)
					checkTags(t, bn.Component.Gateway.InternetGateway.Component.URN(), bn.Component.Gateway.InternetGateway.Component.Tags, tagNamesCheck)
					checkTags(t, bn.Component.Gateway.RouteTable.Component.URN(), bn.Component.Gateway.RouteTable.Component.Tags, tagNamesCheck)
				})
				wg.Wait()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotError := pulumi.RunErr(func(ctx *pulumi.Context) error {
				return CreateRegionComponent(tt.args.name, tt.args.params).GetAndThen(
					ctx,
					func(bn *pgocomp.GetComponentResponse[*RegionComponent]) error {
						tt.test(t, bn)
						return nil
					})
			}, pulumi.WithMocks("project", "stack", mocks(0)))
			if tt.wantError && gotError == nil {
				t.Errorf("NewBasicNetworkComponent() = %v, want error %v", gotError, tt.wantError)
			}
		})
	}
}
