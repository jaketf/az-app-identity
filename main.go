package main

// Credentials for authing requests to azure inferred from following environment variables:
// export AZURE_TENANT_ID="<active_directory_tenant_id"
// export AZURE_CLIENT_ID="<service_principal_appid>"
// export AZURE_CLIENT_SECRET="<service_principal_password>"
// export AZURE_SUBSCRIPTION_ID="<subscription_id>"
// you can grab all of this info from an azure service principal artifact in massdriver

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/manicminer/hamilton/msgraph"
)

var (
	AzureSubscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	AzureTenantID = os.Getenv("AZURE_TEANT_ID")
)


// ApplicationCreate creates an azure cloud applicaiton resource to
// represent the kubenertes app
func ApplicationCreate(ctx context.Context, azCreds *azidentity.DefaultAzureCredential, name string) (graphrbac.Application, error) {
  ac := graphrbac.NewApplicationsClient(AzureTenantID)
  return ac.Create(ctx, graphrbac.ApplicationCreateParameters{
	DisplayName: &name,
  })
}

// ServicePrincipalCreate creates an azure identity for this application
// that can be assigned access policies on azure cloud resources
func ServicePrincipalCreate(ctx context.Context, azCreds *azidentity.DefaultAzureCredential, app *graphrbac.Application) (graphrbac.ServicePrincipal, error) {
  c := graphrbac.NewServicePrincipalsClient(AzureTenantID)
  return c.Create(ctx, graphrbac.ServicePrincipalCreateParameters{
	AppID: app.AppID,
  })
}

// ServicePrincipalCreate creates an azure identity for this application
// that can be assigned access policies on azure cloud resources
func ServicePrincipalPasswordCreate(ctx context.Context, azCreds *azidentity.DefaultAzureCredential, sp *graphrbac.ServicePrincipal) (*msgraph.PasswordCredential, error) {
  c := msgraph.NewServicePrincipalsClient(AzureSubscriptionID) 
  newCredential, _, err := c.AddPassword(ctx, *sp.AppID, msgraph.PasswordCredential{
  })
  return newCredential, err
}

// AddAccessPoliciesToServicePrincipal adds the access policies
// from the connection to this azure service principal to give
// it accesss to the azure cloud resources the app needs to connect to
func AddAccessPoliciesToServicePrincipal(ctx context.Context, azCreds *azidentity.DefaultAzureCredential, sp *graphrbac.ServicePrincipal, policy interface{}) error {
  // TODO
  return nil
}


// TODO pivot to this in the future not yet really stable on azure side for now going to use long lived credential.
// // FederatedIdentityCredentialCreate creates the trust relationship
// // between the KSA used to run the app and the service principal 
// // which has access to the azure cloud resources
// func FederatedIdentityCredentialCreate(ctx context.Context, azCreds *azidentity.DefaultAzureCredential, principal interface{}) error {
//   // TODO 
//   return nil
// }


func main() {
  // Check for subscription id info this would come from service principal artifact
  if len(AzureSubscriptionID) == 0 {
	log.Fatalf("AZURE_SUBSCRIPTION_ID is not set")
  }

  log.Default().Printf("createing azure applicaiton identity resources in subscription id %v\n", AzureSubscriptionID)

  // Create default credentials from environment variables
  appName := "foo"
  ctx := context.Background()
  azCreds, err := azidentity.NewDefaultAzureCredential(nil)
  if err != nil {
	  log.Fatalf("failed to obtain an azurecredential: %v", err)
  }
  log.Default().Printf("azure credential obtained: %#v\n", azCreds)
  policies := []interface{}{} // TODO fill this in with placeholder for what would come from connections

  // All the above will be replaced with wiring into mx provider

  app, err := ApplicationCreate(ctx, azCreds, appName)
  if err != nil {
	log.Fatalf("failed to create application: %v", err)
  }
  sp, err := ServicePrincipalCreate(ctx, azCreds, &app)
  if err != nil {
	log.Fatalf("failed to create service principal: %v", err)
  }
  spPass, err := ServicePrincipalPasswordCreate(ctx, azCreds, &sp)
  log.Default().Printf("service principal password: %#v\n", spPass)
  if err != nil {
	log.Fatalf("failed to create service principal password: %v", err)
 }
	
  for pol := range policies {
	AddAccessPoliciesToServicePrincipal(ctx, azCreds, &sp, pol)
  }

  // these environment variables can be set in the pods that need access to cloud services via this service principal
  // azure has not yet implemented stable workload identity so we are going to use the long lived credential for now
  azCredEnv := map[string]string {
	"AZURE_CLIENT_ID": *sp.AppID,
	"AZURE_CLIENT_SECRET": *spPass.SecretText,
	"AZURE_TENANT_ID": AzureTenantID,
  }
  log.Default().Printf("success! created azure applicaiton identity resources you can use them with this env: %#v", azCredEnv)
}

