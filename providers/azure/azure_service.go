// Copyright 2019 The Terraformer Authors.
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

package azure

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type AzureService struct { //nolint
	terraformutils.Service
}

func (s AzureService) GetSubscriptionID() string {
	return s.Args["config"].(authentication.Config).SubscriptionID
}

func (s AzureService) GetAuthorizer() autorest.Authorizer {
	return s.Args["authorizer"].(autorest.Authorizer)
}
