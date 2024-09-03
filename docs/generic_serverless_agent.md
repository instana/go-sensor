## Generic Serverless Agent

To monitor Go applications deployed in a serverless environment like AWS Lambda, or on a server without a host agent, the process is similar to monitoring any other application. Simply instrument your application with the Instana Go Tracer SDK, deploy it to the appropriate environment, and ensure that the following two environment variables are set.

> **INSTANA_ENDPOINT_URL** - The Instana backend endpoint that your serverless agents connect to. It depends on your region and is different from the host agent backend endpoint.
> **INSTANA_AGENT_KEY** - Your Instana Agent key. The same agent key can be used for host agents and serverless monitoring.

Please note that, in this generic serverless agent setup, only traces are available, metrics are not. However, for certain specific serverless services like AWS Lambda or Fargate, it is possible to correlate infrastructure and collect metrics as well. For more details, please refer to the documentation [here](https://www.ibm.com/docs/en/instana-observability/current?topic=technologies-monitoring-go#platforms).

