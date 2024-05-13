# Tracer Reports Pipeline For Instana Go Tracer

## Install Tekton pipelines for Tracer Reports Generation

- You will find all the required YAML configurations in this folder, including tasks, pipelines, and Cron Job triggers.

### Prerequisites before applying the YAML files

- You need five secrets to run the Tracer Reports Pipeline successfully:

  1. **GitHub enterprise bot token** - You need a GitHub bot token with write access to the `tracer-reports` repository on the enterprise GitHub..
  2. **GitHub enterprise user email** - Email address of the bot user. This email address will be used to add commits to the `tracer-reports` repository.
  3. **Slack token** - Token to send results to the desired slack channel.
  4. **Slack channel ID** - Slack Channel to send pipeline run results.
  5. **Tracer Reports Repo URL** - Repository URL for the `tracer-reports` repository in the enterprise GitHub.

- Once you have access to the above secrets, replace them in the `secrets.yaml` file.

### Installation

- Once you are ready with the above steps, please use the below command to apply the YAML files.

```sh
sh deploy.sh
```

- Congrats! You have successfully configured the Tracer Reports Pipeline. You will see a daily commit to be pushed to the `tracer-reports` repository once daily at 4:00 PM.
