package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2019-08-01/web"
	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type AppServiceGenerator struct {
	AzureService
}

func (g AppServiceGenerator) listApps() ([]terraformutils.Resource, error) {
	AppServiceClient := web.NewAppsClient(g.Args["config"].(authentication.Config).SubscriptionID)
	AppServiceClient.Authorizer = g.Args["authorizer"].(autorest.Authorizer)
	var resources []terraformutils.Resource
	ctx := context.Background()
	apps, err := AppServiceClient.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, site := range apps.Values() {
		resources = append(resources, terraformutils.NewSimpleResource(
			*site.ID,
			*site.Name,
			"azurerm_app_service",
			g.ProviderName,
			[]string{}))
	}

	return resources, nil
}

func (g *AppServiceGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listApps,
	}

	for _, f := range functions {
		resources, err := f()
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}
