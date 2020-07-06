package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"log"
	"net/url"
)

const (
	blobFormatString = `https://%s.blob.core.windows.net`
)

type StorageBlobGenerator struct {
	AzureService
}

//
func (g StorageBlobGenerator) getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(g.GetSubscriptionID())
	storageAccountsClient.Authorizer = g.GetAuthorizer()

	//storageAccountsClient.AddToUserAgent(config.UserAgent())
	return storageAccountsClient
}

func (g StorageBlobGenerator) getAccountKeys(ctx context.Context, accountName, accountGroupName string) (storage.AccountListKeysResult, error) {
	accountsClient := g.getStorageAccountsClient()
	return accountsClient.ListKeys(ctx, accountGroupName, accountName, "kerb")
}

func (g StorageBlobGenerator) getAccountPrimaryKey(ctx context.Context, accountName, accountGroupName string) string {
	response, err := g.getAccountKeys(ctx, accountName, accountGroupName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*response.Keys)[0]).Value)
}

func (g StorageBlobGenerator) getContainerURL(ctx context.Context, accountName, accountGroupName, containerName string) azblob.ContainerURL {
	key := g.getAccountPrimaryKey(ctx, accountName, accountGroupName)
	c, _ := azblob.NewSharedKeyCredential(accountName, key)
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	//{
	//	Telemetry: azblob.TelemetryOptions{Value: config.UserAgent()},
	//})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, accountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

func (g StorageBlobGenerator) ListBlobs(ctx context.Context, accountName, accountGroupName, containerName string) (*azblob.ListBlobsFlatSegmentResponse, error) {
	c := g.getContainerURL(ctx, accountName, accountGroupName, containerName)
	return c.ListBlobsFlatSegment(
		ctx,
		azblob.Marker{},
		azblob.ListBlobsSegmentOptions{
			Details: azblob.BlobListingDetails{
				Snapshots: true,
			},
		})
}

//
func (g StorageBlobGenerator) listStorageBlobs() ([]terraformutils.Resource, error) {
	//var storageBlobs []terraformutils.Resource
	//ctx := context.Background()
	//storageAccountClient := g.getStorageAccountsClient()
	//accountListResult, err := storageAccountClient.List(ctx)
	//if err != nil {
	//	return storageBlobs, err
	//}

	//for _, storageAccount := range *accountListResult.Value {
	//	accountID := *storageAccount.ID
	//	parsedAccountID, err := ParseAzureResourceID(accountID)
	//
	//	blobList := g.ListBlobs(ctx, *storageAccount.Name, parsedAccountID.ResourceGroup)
	//}

	var storageBlobs []terraformutils.Resource
	ctx := context.Background()
	blobContainersClient := storage.NewBlobContainersClient(g.GetSubscriptionID())
	blobContainersClient.Authorizer = g.GetAuthorizer()

	storageAccountGenerator := NewStorageAccountGenerator(g.GetSubscriptionID(), g.GetAuthorizer())
	storageAccountsIterator, err := storageAccountGenerator.GetStorageAccountsIterator()
	if err != nil {
		return storageBlobs, err
	}

	for storageAccountsIterator.NotDone() {
		storageAccount := storageAccountsIterator.Value()
		resourceID, err := ParseAzureResourceID(*storageAccount.ID)
		if err != nil {
			return storageBlobs, err
		}
		blobsForGroupIterator, err := blobContainersClient.ListComplete(ctx, resourceID.ResourceGroup, *storageAccount.Name, "", "")
		if err != nil {
			return storageBlobs, err
		}

		for blobsForGroupIterator.NotDone() {
			blobContainer := blobsForGroupIterator.Value()
			//containerDetails, err := blobContainersClient.Get(ctx, resourceID.ResourceGroup, *storageAccount.Name, *blobContainer.Name)
			//
			//if err != nil {
			//	log.Println(err)
			//}

			blobList, err := g.ListBlobs(ctx, *storageAccount.Name, resourceID.ResourceGroup, *blobContainer.Name)

			if err != nil {
				log.Println(err)
			}

			log.Printf("%v", blobList)

			storageBlobs = append(storageBlobs, terraformutils.NewSimpleResource(
				*blobContainer.ID,
				*blobContainer.Name,
				"azurerm_storage_container",
				"azurerm",
				[]string{}))

			if err := blobsForGroupIterator.NextWithContext(ctx); err != nil {
				log.Println(err)
				break
			}
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
