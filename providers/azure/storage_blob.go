package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"log"
)

type StorageBlobGenerator struct {
	AzureService
}

func (g StorageBlobGenerator) listStorageBlobs() ([]terraformutils.Resource, error) {
	var storageBlobs []terraformutils.Resource
	ctx := context.Background()
	blobStorageClient := storage.NewBlobContainersClient(g.GetSubscriptionID())
	blobStorageClient.SubscriptionID = g.GetSubscriptionID()
	blobStorageClient.Authorizer = g.GetAuthorizer()

	storageAccountGenerator := NewStorageAccountGenerator(g.GetSubscriptionID(), g.GetAuthorizer())
	storageAccountsIterator, err := storageAccountGenerator.GetStorageAccountsIterator()
	if err != nil {
		return storageBlobs, err
	}

	for storageAccountsIterator.NotDone() {
		account := storageAccountsIterator.Value()
		resourceID, err := ParseAzureResourceID(*account.ID)
		if err != nil {
			return storageBlobs, err
		}
		blobsForGroupIterator, err := blobStorageClient.ListComplete(ctx, resourceID.ResourceGroup, *account.Name, "", "")
		if err != nil {
			return storageBlobs, err
		}

		blobsForGroup := blobsForGroupIterator.Values()

		for _, blob := range blobsForGroup {
			storageBlobs = append(storageBlobs, terraformutils.NewSimpleResource(
				*blob.ID,
				*blob.Name,
				"azurerm_storage_blob",
				g.ProviderName,
				[]string{}))
		}

		if err := storageAccountsIterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			break
		}
	}

	return storageBlobs, nil
}

func (g *StorageBlobGenerator) InitResources() error {
	resources, err := g.listStorageBlobs()
	if err != nil {
		return err
	}

	g.Resources = append(g.Resources, resources...)

	return nil
}
