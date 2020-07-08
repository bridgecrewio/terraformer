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
	"log"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type StorageContainerGenerator struct {
	AzureService
}

func (g StorageContainerGenerator) listBlobContainers(accountListResultIterator storage.AccountListResultIterator) []terraformutils.Resource {
	var containerResources []terraformutils.Resource
	blobServiceClient := storage.NewBlobContainersClient(g.GetSubscriptionID())
	blobServiceClient.Authorizer = g.GetAuthorizer()
	ctx := context.Background()

	for accountListResultIterator.NotDone() {
		storageAccount := accountListResultIterator.Value()
		parsedAccountID, err := ParseAzureResourceID(*storageAccount.ID)
		if err != nil {
			log.Println(err)
			break
		}
		containerItemsIterator, err := blobServiceClient.ListComplete(ctx, parsedAccountID.ResourceGroup, *storageAccount.Name, "", "", "")
		if err != nil {
			log.Println(err)
			break
		}

		for containerItemsIterator.NotDone() {
			containerItem := containerItemsIterator.Value()
			containerResources = append(containerResources, terraformutils.NewSimpleResource(
				*containerItem.ID,
				*storageAccount.Name+"-"+*containerItem.Name,
				"azurerm_storage_container",
				"azurerm",
				[]string{}))

			log.Println(containerItem)
			if err := containerItemsIterator.NextWithContext(ctx); err != nil {
				log.Println(err)
				break
			}
		}

		if err := accountListResultIterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			break
		}
	}

	return containerResources
}

func (g *StorageContainerGenerator) GetStorageAccountsIterator() (storage.AccountListResultIterator, error) {
	ctx := context.Background()
	accountsClient := storage.NewAccountsClient(g.GetSubscriptionID())

	accountsClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)
	accountsIterator, err := accountsClient.ListComplete(ctx)

	return accountsIterator, err
}

func (g *StorageContainerGenerator) InitResources() error {
	accountsIterator, err := g.GetStorageAccountsIterator()
	if err != nil {
		return err
	}

	storageAccounts := g.listBlobContainers(accountsIterator)
	g.Resources = storageAccounts

	return nil
}
