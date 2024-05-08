# Tracer Report Pipeline For Instana Go Tracer

## Install Tekton pipelines for Tracer Reports Generation

- You will find all the required YAML configurations in this folder. This includes all the required tasks, pipelines, and Cron Job triggers, etc.

### Prerequisites before applying the YAML files

- You need five secrets to run the CI Build pipeline successfully:

  1. **GitHub enterprise bot token** - You need a GitHub bot token with write access to the `tracer-reports` repo in the enterprise GitHub.
  2. **GitHub enterprise user email** - Email ID of the bot user. This email ID will be used to add commits to `tracer-reports` repo.
  2. **Slack token** - Token to send result to the desired channel.
  3. **Slack channel ID** - Channel to send pipeline run result.
  2. **Tracer Reports Repo URL** - Repo URL for `tracer-reports` repo, in the enterprise GitHub.

- Once you have access to the above secrets, replace them in the `secrets.yaml` file.

### Installation

- Once you are ready with the above steps, please use the below command to apply the YAML files.

```sh
sh deploy.sh
```

- Congrats! You have successfully configured Tracer Report Pipeline. You will see a commit to be pushed to `tracer-reports` branch daily once (At 04:00 PM).
