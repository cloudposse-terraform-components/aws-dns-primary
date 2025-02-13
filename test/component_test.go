package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cloudposse/test-helpers/pkg/atmos"
	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/gruntwork-io/terratest/modules/random"
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

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "dns-primary/basic"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	randomID := strings.ToLower(random.UniqueId())
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

	defer s.DestroyAtmosComponent(s.T(), component, stack, &inputs)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, &inputs)
	assert.NotNil(s.T(), options)

	zones := map[string]zone{}
	atmos.OutputStruct(s.T(), options, "zones", &zones)
	zone := zones[domainName]

	DomainRecordName := fmt.Sprintf("%s.", domainName)
	aRecord := aws.GetRoute53Record(s.T(), zone.ZoneID, zone.Name, "A", awsRegion)
	assert.Equal(s.T(), DomainRecordName, *aRecord.Name)
	assert.EqualValues(s.T(), 60, *aRecord.TTL)
	assert.Equal(s.T(), "127.0.0.1", *aRecord.ResourceRecords[0].Value)

	wwwDomain := fmt.Sprintf("www.%s", domainName)
	wwwDomainName := fmt.Sprintf("%s.", wwwDomain)
	wwwRecord := aws.GetRoute53Record(s.T(), zone.ZoneID, wwwDomain, "CNAME", awsRegion)
	assert.Equal(s.T(), wwwDomainName, *wwwRecord.Name)
	assert.EqualValues(s.T(), 60, *wwwRecord.TTL)
	assert.Equal(s.T(), domainName, *wwwRecord.ResourceRecords[0].Value)

	cNameDomain := fmt.Sprintf("123456.%s", domainName)
	cNameDomainName := fmt.Sprintf("%s.", cNameDomain)
	cNameRecord := aws.GetRoute53Record(s.T(), zone.ZoneID, cNameDomain, "CNAME", awsRegion)
	assert.Equal(s.T(), cNameDomainName, *cNameRecord.Name)
	assert.EqualValues(s.T(), 120, *cNameRecord.TTL)
	assert.Equal(s.T(), domainName, *cNameRecord.ResourceRecords[0].Value)

	s.DriftTest(component, stack, &inputs)
}

func (s *ComponentSuite) TestEnabledFlag() {
	const component = "dns-primary/disabled"
	const stack = "default-test"

	randomID := strings.ToLower(random.UniqueId())
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

	s.VerifyEnabledFlag(component, stack, &inputs)
}


func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)
	helper.Run(t, suite)
}
