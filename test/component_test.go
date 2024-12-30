package test

import (
	"fmt"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/aws-component-helper"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
)

type zone struct {
	Arn               string            `json:"arn"`
	Comment           string            `json:"comment"`
	DelegationSetId   string            `json:"delegation_set_id"`
	ForceDestroy      bool              `json:"force_destroy"`
	Id                string            `json:"id"`
	Name              string            `json:"name"`
	NameServers       []string          `json:"name_servers"`
	PrimaryNameServer string            `json:"primary_name_server"`
	Tags              map[string]string `json:"tags"`
	TagsAll           map[string]string `json:"tags_all"`
	Vpc               []struct {
		ID     string `json:"vpc_id"`
		Region string `json:"vpc_region"`
	} `json:"vpc"`
	ZoneID string `json:"zone_id"`
}

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
						"ttl":       120,
						"records":   []string{domainName},
					},
				},
			}

			defer atm.GetAndDestroy("dns-primary/basic", "default-test", inputs)
			component := atm.GetAndDeploy("dns-primary/basic", "default-test", inputs)

			zones := map[string]zone{}
			atm.OutputStruct(component, "zones", &zones)
			zone := zones[domainName]

			DomainRecordName := fmt.Sprintf("%s.", domainName)
			aRecord := aws.GetRoute53Record(t, zone.ZoneID, zone.Name, "A", awsRegion)
			assert.Equal(t, DomainRecordName, *aRecord.Name)
			assert.EqualValues(t, 60, *aRecord.TTL)
			assert.Equal(t, "127.0.0.1", *aRecord.ResourceRecords[0].Value)

			wwwDomain := fmt.Sprintf("www.%s", domainName)
			wwwDomainName := fmt.Sprintf("%s.", wwwDomain)
			wwwRecord := aws.GetRoute53Record(t, zone.ZoneID, wwwDomain, "CNAME", awsRegion)
			assert.Equal(t, wwwDomainName, *wwwRecord.Name)
			assert.EqualValues(t, 60, *wwwRecord.TTL)
			assert.Equal(t, domainName, *wwwRecord.ResourceRecords[0].Value)

			cNameDomain := fmt.Sprintf("123456.%s", domainName)
			cNameDomainName := fmt.Sprintf("%s.", cNameDomain)
			cNameRecord := aws.GetRoute53Record(t, zone.ZoneID, cNameDomain, "CNAME", awsRegion)
			assert.Equal(t, cNameDomainName, *cNameRecord.Name)
			assert.EqualValues(t, 120, *cNameRecord.TTL)
			assert.Equal(t, domainName, *cNameRecord.ResourceRecords[0].Value)
		})
	})
}
