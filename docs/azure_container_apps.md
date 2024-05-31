## How to Monitor Serverless Go Application Deployed in Container Apps

To monitor a Go application deployed in Azure Container Apps, you can instrument the application with the Instana Go Tracer SDK and deploy it to the container apps similar to other platforms. Additionally, ensure the following environment variables are set in Container Apps to achieve proper infrastructure correlation:

> **AZURE_SUBSCRIPTION_ID** = <[azure_subscription_id](https://learn.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-subscription)>
> **AZURE_RESOURCE_GROUP** = <[azure_resource_group_name](https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/manage-resource-groups-portal#what-is-a-resource-group)>

These environment variables are essential for constructing your container apps' resource ID. The Azure Container Apps resource ID follows this format: 

*`/subscriptions/{AZURE_SUBSCRIPTION_ID}/resourceGroups/{AZURE_RESOURCE_GROUP}/providers/Microsoft.App/containerapps/<container_app_name>`*.