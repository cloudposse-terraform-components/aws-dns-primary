package test

import (
	"fmt"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	awsRegion := "us-east-2"

	fixture := helper.NewFixture(t, "../", awsRegion, "test/fixtures")

	defer fixture.TearDown()
	fixture.SetUp(&atmos.Options{})

	fixture.Suite("default", func(t *testing.T, suite *helper.Suite) {
		suite.Test(t, "basic", func(t *testing.T, atm *helper.Atmos) {
			randomID := suite.GetRandomIdentifier()
			domainName := fmt.Sprintf("example-%s.net", randomID)
			inputs := map[string]interface{}{
				"domain_names": []string{domainName},
				"record_config": []map[string]interface{}{
					{
						"root_zone": domainName,
						"name":      "",
						"type":      "A",
						"ttl":       60,
						"records":   []string{"127.0.0.1"},
					},
					{
						"root_zone": domainName,
						"name":      "www.",
						"type":      "CNAME",
						"ttl":       60,
						"records":   []string{domainName},
					},
					{
						"root_zone": domainName,
						"name":      "123456.",
						"type":      "CNAME",
						"ttl":       60,
						"records":   []string{domainName},
					},
				},
			}

			defer atm.GetAndDestroy("dns-primary/basic", "default-test", inputs)
			component := atm.GetAndDeploy("dns-primary/basic", "default-test", inputs)

			outputs := atm.OutputAll(component)
			require.Equal(t, "", outputs)

			// vpcId := atm.Output(component, "vpc_id")
			// require.True(t, strings.HasPrefix(vpcId, "vpc-"))

			// vpc := aws.GetVpcById(t, vpcId, awsRegion)

			// assert.Equal(t, vpc.Name, fmt.Sprintf("eg-default-ue2-test-vpc-terraform-%s", component.RandomIdentifier))
			// assert.Equal(t, *vpc.CidrAssociations[0], "172.16.0.0/16")
			// assert.Equal(t, *vpc.CidrBlock, "172.16.0.0/16")
			// assert.Nil(t, vpc.Ipv6CidrAssociations)
			// assert.Equal(t, vpc.Tags["Environment"], "ue2")
			// assert.Equal(t, vpc.Tags["Namespace"], "eg")
			// assert.Equal(t, vpc.Tags["Stage"], "test")
			// assert.Equal(t, vpc.Tags["Tenant"], "default")

			// subnets := vpc.Subnets
			// require.Equal(t, 2, len(subnets))

			// public_subnet_ids := atm.OutputList(component, "public_subnet_ids")
			// assert.Empty(t, public_subnet_ids)

			// public_subnet_cidrs := atm.OutputList(component, "public_subnet_cidrs")
			// assert.Empty(t, public_subnet_cidrs)

			// private_subnet_ids := atm.OutputList(component, "private_subnet_ids")
			// assert.Equal(t, 2, len(private_subnet_ids))

			// assert.Contains(t, private_subnet_ids, subnets[0].Id)
			// assert.Contains(t, private_subnet_ids, subnets[1].Id)

			// assert.False(t, aws.IsPublicSubnet(t, subnets[0].Id, awsRegion))
			// assert.False(t, aws.IsPublicSubnet(t, subnets[1].Id, awsRegion))

			// nats, err := GetNatsByVpcIdE(t, vpcId, awsRegion)
			// assert.NoError(t, err)
			// assert.Equal(t, 0, len(nats))
		})
	})
}
