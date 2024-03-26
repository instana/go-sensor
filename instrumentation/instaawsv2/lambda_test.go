// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	instana "github.com/instana/go-sensor"
	instaawsv2 "github.com/instana/go-sensor/instrumentation/instaawsv2"
	"github.com/stretchr/testify/assert"
)

func TestInvokeLambda(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ctx := context.Background()
	ps := sensor.Tracer().StartSpan("aws-lambda-parent-span")
	ctx = instana.ContextWithSpan(ctx, ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	lambdaClient := lambda.NewFromConfig(cfg)

	lambdaFuncName := "lambda-function-name"
	_, err = lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: &lambdaFuncName,
	})
	assert.NoError(t, err, "Error while invoking the lambda")

	ps.Finish()

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recorderSpans))

	lambdaSpan := recorderSpans[0]
	assert.IsType(t, instana.AWSLambdaInvokeSpanData{}, lambdaSpan.Data)

	data := lambdaSpan.Data.(instana.AWSLambdaInvokeSpanData)
	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   lambdaFuncName,
		InvocationType: string(types.InvocationTypeRequestResponse),
		Error:          "",
	}, data.Tags)
}

func TestInvokeLambdaNoParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	lambdaClient := lambda.NewFromConfig(cfg)

	lambdaFuncName := "lambda-function-name"
	_, err = lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: &lambdaFuncName,
	})
	assert.NoError(t, err, "Error while invoking the lambda")

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 1, len(recorderSpans))
}

func TestInvokeLambdaWithClientContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ctx := context.Background()
	ps := sensor.Tracer().StartSpan("aws-lambda-parent-span")
	ctx = instana.ContextWithSpan(ctx, ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	lambdaClient := lambda.NewFromConfig(cfg)

	lambdaFuncName := "lambda-function-name"
	clientCtxEncoded := "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiIiwiYXBwX3RpdGxlIjoiIiwiYXBwX3ZlcnNpb25fY29kZSI6IiI" +
		"sImFwcF9wYWNrYWdlX25hbWUiOiIifSwiZW52IjpudWxsLCJjdXN0b20iOnsieC1pbnN0YW5hLWwiOiIxIiwieC1pbnN0YW5hLXMiOiI2ZmU" +
		"wN2IwYzkzYjJmYTk5IiwieC1pbnN0YW5hLXQiOiIyZTMxM2FmZjZjNmEyMmQyIn19Cg=="

	_, err = lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:  &lambdaFuncName,
		ClientContext: &clientCtxEncoded,
	})
	assert.NoError(t, err, "Error while invoking the lambda")

	ps.Finish()

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recorderSpans))

	lambdaSpan := recorderSpans[0]
	assert.IsType(t, instana.AWSLambdaInvokeSpanData{}, lambdaSpan.Data)

	data := lambdaSpan.Data.(instana.AWSLambdaInvokeSpanData)
	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   lambdaFuncName,
		InvocationType: string(types.InvocationTypeRequestResponse),
		Error:          "",
	}, data.Tags)
}

func TestInvokeLambdaWithInvalidClientContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ctx := context.Background()
	ps := sensor.Tracer().StartSpan("aws-lambda-parent-span")
	ctx = instana.ContextWithSpan(ctx, ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	lambdaClient := lambda.NewFromConfig(cfg)

	lambdaFuncName := "lambda-function-name"
	clientCtxEncoded := testString(30)

	_, err = lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:  &lambdaFuncName,
		ClientContext: clientCtxEncoded,
	})
	assert.NoError(t, err, "Error while invoking the lambda")

	ps.Finish()

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recorderSpans))

	lambdaSpan := recorderSpans[0]
	assert.IsType(t, instana.AWSLambdaInvokeSpanData{}, lambdaSpan.Data)

	data := lambdaSpan.Data.(instana.AWSLambdaInvokeSpanData)
	assert.Equal(t, instana.AWSInvokeSpanTags{
		FunctionName:   lambdaFuncName,
		InvocationType: string(types.InvocationTypeRequestResponse),
		Error:          "",
	}, data.Tags)
}
