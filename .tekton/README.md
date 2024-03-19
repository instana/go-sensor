# Tekton CI for Instana Go Tracer

## Tekton Setup on a Cluster

- You need access to a cluster with full admin privileges.
- Allocate enough RAM and CPU so that all the pods, including sidecar pods, will run smoothly on a single node.
- Add multiple nodes to increase parallel pipeline runs.

### Tekton Setup

```sh
# Install Tekton pipelines
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

# Install Tekton triggers
kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml

# Install Tekton interceptors
kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml

# Install Tekton dashboard - full installation is needed for read/write capabilities. eg: to make changes in the pipeline, such as re-running a pipeline run or deleting a pipeline run. 
kubectl apply --filename https://storage.googleapis.com/tekton-releases/dashboard/latest/release-full.yaml

# Make sure all pods are in the ready state before proceeding further by issuing the following command.
kubectl get pods --namespace tekton-pipelines --watch

# To access the dashboard in localhost
kubectl proxy
```

- If you have successfully completed the above mentioned steps, you should be able to access the Tekton Dashboard from [here](http://localhost:8001/api/v1/namespaces/tekton-pipelines/services/tekton-dashboard:http/proxy/)


## Install Tekton pipelines for Go Tracer
- You will find all the required YAML configurations in the `.tekton` folder, present in the root directory of the go-tracer repo. This includes all the required tasks, pipelines, and GitHub triggers, etc.

### Prerequisites before applying the YAML files

- You need to create three secrets to run the Go tracer pipelines:
    1. **GitHub bot token** - You need a GitHub bot token with write access to the repo. This is for sending commit statuses.
    2. **GitHub Webhook Secret** - Create a very long random secret. You need to add this to the GitHub UI when creating a webhook for PR events.
    3. **Cosmos URL and Secret** - This is for running azcosmos integration tests.

- Once you have access to the above secrets, please replace them in the `secrets.yaml` file.
- You need an ingress controller for the GitHub Webhook to come through.
- Please replace the `ingressClassName` and ingress domain or subdomain URL in the `github-webhook-ingress.yaml` file.
- Make sure you create a GitHub webhook for the `Pull requests` events in the settings tab of the repo. Please add the previously created webhook secret and `<<ingress_url/hooks>>` as the Payload URL in the appropriate place when creating the webhook.

### Installation
- Once you are ready with the above steps, please use the below command to apply the YAML files.
```sh
sh deploy.sh
```
- Congrats! You have successfully configured everything. You will see a status has been posted for the Tekton runs whenever a new PR gets created.

## How to debug/re-run a pipeline run

- You will find the Tekton dashboard URL for a specific pipeline run from the `details` section of the commit status.
-You can access the dashboard once you have set up and authenticated the cluster in your local and using the `kubectl proxy` command.  How to do this using IBM Cloud, you can see the documentation [here](https://cloud.ibm.com/docs/containers?topic=containers-access_cluster#access_public_se) for help.
- You can see the logs or re-run the pipeline run once you have access to the dashboard.
- You will see that the statuses in the PR will update once you initiate a re-run.

## Helpful resources
- [Ingress in IBM Cloud](https://cloud.ibm.com/docs/containers?topic=containers-managed-ingress-about)
- [Tekton: Getting Started](https://tekton.dev/docs/getting-started/)
- [Accessing clusters through the public cloud service endpoint on ibm cloud](https://cloud.ibm.com/docs/containers?topic=containers-access_cluster#access_public_se)
- [Creating webhooks on Github](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks)
- [Create a commit status](https://docs.github.com/en/rest/commits/statuses?apiVersion=2022-11-28#create-a-commit-status)
