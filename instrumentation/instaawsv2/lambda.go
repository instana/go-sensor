// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

const maxClientContextLen = 3582

type AWSInvokeLambdaOperations struct{}

var _ AWSOperations = (*AWSInvokeLambdaOperations)(nil)

func (o AWSInvokeLambdaOperations) injectContextWithSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	// An exit span will be created independently without a parent span
	// and sent if the user has opted in.
	opts := []ot.StartSpanOption{}
	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		opts = append(opts, ot.ChildOf(parent.Context()))
	}
	sp := tr.Tracer().StartSpan("aws.lambda.invoke", opts...)

	if invokeInputReq, ok := params.(*lambda.InvokeInput); ok {

		lambdaFuncName := stringDeRef(invokeInputReq.FunctionName)
		sp.SetTag(lambdaFunction, lambdaFuncName)

		invocationType := invokeInputReq.InvocationType
		if invocationType == "" {
			invocationType = types.InvocationTypeRequestResponse
		}
		sp.SetTag(typeTag, string(invocationType))

		if err := o.injectSpanToCarrier(params, sp); err != nil {
			tr.Logger().Error("failed to inject lambda span to carrier.")
		}
	}

	return instana.ContextWithSpan(ctx, sp)
}

func (o AWSInvokeLambdaOperations) finishSpan(tr instana.TracerLogger, ctx context.Context, err error) {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the sqs child span from context.")
		return
	}
	defer sp.Finish()

	if err != nil {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(errorTag, err.Error())
	}
}

func (o AWSInvokeLambdaOperations) injectSpanToCarrier(params interface{}, sp ot.Span) error {
	var p *lambda.InvokeInput
	var ok bool

	if p, ok = params.(*lambda.InvokeInput); !ok {
		return fmt.Errorf("received params is not of type lambda.InvokeInput")
	}

	var err error
	lc := lambdaClientContext{}

	if p.ClientContext != nil {
		lc, err = newLambdaClientContextFromBase64EncodedJSON(*p.ClientContext)
		if err != nil {
			return errors.Wrap(err, "failed to create lambda ClientContext")
		}
	}

	if lc.Custom == nil {
		lc.Custom = make(map[string]string)
	}

	sp.Tracer().Inject(
		sp.Context(),
		ot.TextMap,
		ot.TextMapCarrier(lc.Custom),
	)

	var encodedStr string
	encodedStr, err = lc.base64JSON()
	if err != nil {
		return errors.Wrap(err, "failed to marshall the ClientContext")
	}

	if len(encodedStr) <= maxClientContextLen {
		p.ClientContext = &encodedStr
	}

	return nil
}

func (o AWSInvokeLambdaOperations) extractTags(interface{}) (ot.Tags, error) {
	return ot.Tags{}, nil
}

// lambdaClientContextClientApplication represent client application specific data part of the lambdaClientContext.
type lambdaClientContextClientApplication struct {
	InstallationID string `json:"installation_id"`
	AppTitle       string `json:"app_title"`
	AppVersionCode string `json:"app_version_code"`
	AppPackageName string `json:"app_package_name"`
}

// lambdaClientContext represents ClientContext from the AWS Invoke call https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax.
type lambdaClientContext struct {
	Client lambdaClientContextClientApplication
	Env    map[string]string `json:"env"`
	Custom map[string]string `json:"custom"`
}

// base64JSON marshal lambdaClientContext to JSON and returns it as the base64 encoded string or error if any occurs.
func (lc *lambdaClientContext) base64JSON() (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	if err := json.NewEncoder(encoder).Encode(*lc); err != nil {
		return "", fmt.Errorf("lambda client context encoder encode: %v", err.Error())
	}

	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("lambda client context encoder close: %v", err.Error())
	}

	return buf.String(), nil
}

// newLambdaClientContextFromBase64EncodedJSON creates lambdaClientContext from the base64 encoded JSON or returns
// error if there is decoding error.
func newLambdaClientContextFromBase64EncodedJSON(data string) (lambdaClientContext, error) {
	reader := strings.NewReader(data)
	decoder := base64.NewDecoder(base64.StdEncoding, reader)

	res := lambdaClientContext{}
	if err := json.NewDecoder(decoder).Decode(&res); err != nil {
		return res, fmt.Errorf("can't decode lambda client context: %v", err.Error())
	}

	return res, nil
}
