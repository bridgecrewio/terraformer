// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"strconv"
)

var CloudWatchLogsAllowEmptyValues = []string{"tags.", "retention_in_days"}

type CloudWatchLogsGenerator struct {
	AWSService
}

func (g *CloudWatchLogsGenerator) createResources(config aws.Config, logGroups *cloudwatchlogs.DescribeLogGroupsResponse, region string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	//svc := cloudwatchlogs.New(config)
	for _, logGroup := range logGroups.LogGroups {
		resourceName := aws.StringValue(logGroup.LogGroupName)

		attributes := map[string]string{}

		if logGroup.RetentionInDays != nil {
			attributes["retention_in_days"] = strconv.FormatInt(*logGroup.RetentionInDays, 10)
		}

		if logGroup.KmsKeyId != nil {
			attributes["kms_key_id"] = *logGroup.KmsKeyId
		}

		resources = append(resources, terraformutils.NewResource(
			resourceName,
			resourceName,
			"aws_cloudwatch_log_group",
			"aws",
			attributes,
			CloudWatchLogsAllowEmptyValues,
			map[string]interface{}{}))
	}
	return resources
}

// Generate TerraformResources from AWS API
func (g *CloudWatchLogsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cloudwatchlogs.New(config)

	logGroups, err := svc.DescribeLogGroupsRequest(&cloudwatchlogs.DescribeLogGroupsInput{}).Send(context.Background())
	if err != nil {
		return err
	}
	g.Resources = g.createResources(config, logGroups, g.GetArgs()["region"].(string))
	return nil
}

// remove retention_in_days if it is 0 (it gets added by the "refresh" stage)
func (g *CloudWatchLogsGenerator) PostConvertHook() error {
	for _, resource := range g.Resources {
		if resource.Item["retention_in_days"] == "0" {
			delete(resource.Item, "retention_in_days")
		}
	}
	return nil
}
