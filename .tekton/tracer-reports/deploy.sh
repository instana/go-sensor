#!/bin/sh
# (c) Copyright IBM Corp. 2024

kubectl apply --filename secrets.yaml
kubectl apply --filename task.yaml
kubectl apply --filename pipeline.yaml
kubectl apply --filename ../tekton-triggers-eventlistener-serviceaccount.yaml
kubectl apply --filename cron-eventlistener.yaml
kubectl apply --filename cronjob.yaml
kubectl apply --filename ../cleanup-cron-job.yaml
