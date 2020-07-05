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
	"context"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type StorageAccountGenerator struct {
	AzureService
}

func NewStorageAccountGenerator(subscriptionID string, authorizer autorest.Authorizer) *StorageAccountGenerator {
	storageAccountGenerator := new(StorageAccountGenerator)
	storageAccountGenerator.Args = map[string]interface{}{}
	storageAccountGenerator.Args["config"] = authentication.Config{SubscriptionID: subscriptionID}
	storageAccountGenerator.Args["authorizer"] = authorizer

	return storageAccountGenerator
}

func (g StorageAccountGenerator) createResources(accountListResultIterator storage.AccountListResultIterator) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for accountListResultIterator.NotDone() {
		account := accountListResultIterator.Value()
		resources = append(resources, terraformutils.NewSimpleResource(
			*account.ID,
			*account.Name,
			"azurerm_storage_account",
			"azurerm",
			[]string{}))
		if err := accountListResultIterator.Next(); err != nil {
			log.Println(err)
			break
		}
	}
	return resources
}

func (g *StorageAccountGenerator) GetStorageAccountsIterator() (storage.AccountListResultIterator, error) {
	ctx := context.Background()
	accountsClient := storage.NewAccountsClient(g.GetSubscriptionID())

	accountsClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)
	accountsIterator, err := accountsClient.ListComplete(ctx)

	return accountsIterator, err
}

func (g *StorageAccountGenerator) InitResources() error {
	accountsIterator, err := g.GetStorageAccountsIterator()
	if err != nil {
		return err
	}

	storageAccounts := g.createResources(accountsIterator)
	g.Resources = storageAccounts

	return nil
}
