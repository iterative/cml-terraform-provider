package resources

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-iterative/task/az/client"
	"terraform-provider-iterative/task/common"
)

func NewCredentials(client *client.Client, identifier common.Identifier, resourceGroup *ResourceGroup, storageAccount *StorageAccount, blobContainer *BlobContainer) *Credentials {
	c := &Credentials{
		client:     client,
		Identifier: identifier.Long(),
	}
	c.Dependencies.ResourceGroup = resourceGroup
	c.Dependencies.StorageAccount = storageAccount
	c.Dependencies.BlobContainer = blobContainer
	return c
}

type Credentials struct {
	client       *client.Client
	Identifier   string
	Dependencies struct {
		ResourceGroup  *ResourceGroup
		StorageAccount *StorageAccount
		BlobContainer  *BlobContainer
	}
	Resource map[string]string
}

func (c *Credentials) Read(ctx context.Context) error {
	connectionString := fmt.Sprintf(
		":azureblob,account='%s',key='%s':%s",
		c.Dependencies.StorageAccount.Identifier,
		*c.Dependencies.StorageAccount.Attributes.Value,
		c.Dependencies.BlobContainer.Identifier,
	)

	credentials, err := c.client.Settings.GetClientCredentials()
	if err != nil {
		return err
	}

	if len(credentials.ClientSecret) == 0 {
		return errors.New("unable to find client secret")
	}

	subscriptionID := c.client.Settings.GetSubscriptionID()

	c.Resource = map[string]string{
		"AZURE_CLIENT_ID":         credentials.ClientID,
		"AZURE_CLIENT_SECRET":     credentials.ClientSecret,
		"AZURE_SUBSCRIPTION_ID":   subscriptionID,
		"AZURE_TENANT_ID":         credentials.TenantID,
		"RCLONE_REMOTE":           connectionString,
		"TPI_TASK_CLOUD_PROVIDER": string(c.client.Cloud.Provider),
		"TPI_TASK_CLOUD_REGION":   c.client.Region,
		"TPI_TASK_IDENTIFIER":     c.Identifier,
	}

	return nil
}
