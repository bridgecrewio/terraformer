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
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3AllowEmptyValues = []string{"tags."}

var S3AdditionalFields = map[string]interface{}{}

type S3Generator struct {
	AWSService
}

// createResources iterate on all buckets
// for each bucket we check region and choose only bucket from set region
// for each bucket try get bucket policy, if policy exist create additional NewTerraformResource for policy
func (g *S3Generator) createResources(config aws.Config, buckets *s3.ListBucketsResponse, region string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	svc := s3.New(config)
	for _, bucket := range buckets.Buckets {
		resourceName := aws.StringValue(bucket.Name)
		location, err := svc.GetBucketLocationRequest(&s3.GetBucketLocationInput{Bucket: bucket.Name}).Send(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}
		// check if bucket in region
		constraintString, _ := s3.NormalizeBucketLocation(location.LocationConstraint).MarshalValue()
		if constraintString == region {
			attributes := map[string]string{
				"force_destroy": "false",
				"acl":           "private",
			}
			// try get policy
			policy, err := svc.GetBucketPolicyRequest(&s3.GetBucketPolicyInput{
				Bucket: bucket.Name,
			}).Send(context.Background())

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() != "NoSuchBucketPolicy" {
					log.Println(err)
					continue
				}
			} else {
				attributes["policy"] = *policy.Policy
			}
			resources = append(resources, terraformutils.NewResource(
				resourceName,
				resourceName,
				"aws_s3_bucket",
				"aws",
				attributes,
				S3AllowEmptyValues,
				S3AdditionalFields))
		}
	}
	return resources
}

// Generate TerraformResources from AWS API,
// from each s3 bucket create 2 TerraformResource(bucket and bucket policy)
// Need bucket name as ID for terraform resource
func (g *S3Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3.New(config)

	buckets, err := svc.ListBucketsRequest(&s3.ListBucketsInput{}).Send(context.Background())
	if err != nil {
		return err
	}
	g.Resources = g.createResources(config, buckets, g.GetArgs()["region"].(string))
	return nil
}

// PostGenerateHook for add bucket policy json as heredoc
// support only bucket with policy
func (g *S3Generator) PostConvertHook() error {
	for i, resource := range g.Resources {
		if resource.InstanceInfo.Type == "aws_s3_bucket" {
			if val, ok := g.Resources[i].Item["acl"]; ok && val == "private" {
				delete(g.Resources[i].Item, "acl")
			}
			if val, ok := g.Resources[i].Item["policy"]; ok {
				g.Resources[i].Item["policy"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, g.escapeAwsInterpolation(val.(string)))
			}
		}
	}
	return nil
}
