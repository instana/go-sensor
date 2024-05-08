# Tekton Pipelines for Instana Go Tracer

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

## Pipelines present in Go Tracer
- Individual pipelines can be found in their respective directories here. You will find detailed usage instructions in a README file within each directory.

  1. [CI Build](./ci-build/README.md)
  2. [Tracer Reports](./tracer-reports/README.md)

## Deleting old pipeline run resources

- Deletion of old pipeline run resources will be automatically handled by a cron job by default. You can review the configuration in `cleanup-cron-job.yaml`. Feel free to edit the `NUM_TO_KEEP` variable to specify the number of old pipeline runs you wish to retain. The default value is `50`.

## Helpful resources

- [Ingress in IBM Cloud](https://cloud.ibm.com/docs/containers?topic=containers-managed-ingress-about)
- [Tekton: Getting Started](https://tekton.dev/docs/getting-started/)
- [Accessing clusters through the public cloud service endpoint on ibm cloud](https://cloud.ibm.com/docs/containers?topic=containers-access_cluster#access_public_se)
- [Creating webhooks on Github](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks)
- [Create a commit status](https://docs.github.com/en/rest/commits/statuses?apiVersion=2022-11-28#create-a-commit-status)
- [CronJob in Tekton](https://github.com/tektoncd/triggers/tree/main/examples/v1beta1/cron)
