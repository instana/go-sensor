// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownDynamoDBMethod = errors.New("dynamodb method not instrumented")

type AWSDynamoDBOperations struct{}

var _ AWSOperations = (*AWSDynamoDBOperations)(nil)

func (o AWSDynamoDBOperations) injectContextWithSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := o.extractTags(params)
	if err != nil {
		if errors.Is(err, errUnknownDynamoDBMethod) {
			tr.Logger().Error("failed to identify the dynamodb method: ", err.Error())
			return ctx
		}
	}

	// An exit span will be created independently without a parent span
	// and sent if the user has opted in.
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCClient,
		ot.Tags{
			dynamodbRegion: awsmiddleware.GetRegion(ctx),
		},
		tags,
	}
	parent, ok := instana.SpanFromContext(ctx)
	if ok {
		opts = append(opts, ot.ChildOf(parent.Context()))
	}
	sp := tr.Tracer().StartSpan("dynamodb", opts...)

	return instana.ContextWithSpan(ctx, sp)
}

// finishDynamoDBSpan retrieves tags from completed calls and adds them to the span
func (o AWSDynamoDBOperations) finishSpan(tr instana.TracerLogger, ctx context.Context, err error) {
	sp, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the dynamodb child span from context.")
		return
	}
	defer sp.Finish()

	if err != nil {
		sp.LogFields(otlog.Error(err))
		sp.SetTag(dynamodbError, err.Error())
	}
}

func (o AWSDynamoDBOperations) extractTags(params interface{}) (ot.Tags, error) {
	switch params := params.(type) {
	case *dynamodb.CreateTableInput:
		return ot.Tags{
			dynamodbOp:    "create",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.ListTablesInput:
		return ot.Tags{
			dynamodbOp: "list",
		}, nil
	case *dynamodb.GetItemInput:
		return ot.Tags{
			dynamodbOp:    "get",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.PutItemInput:
		return ot.Tags{
			dynamodbOp:    "put",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.UpdateItemInput:
		return ot.Tags{
			dynamodbOp:    "update",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.DeleteItemInput:
		return ot.Tags{
			dynamodbOp:    "delete",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.QueryInput:
		return ot.Tags{
			dynamodbOp:    "query",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.ScanInput:
		return ot.Tags{
			dynamodbOp:    "scan",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	default:
		return nil, errUnknownDynamoDBMethod
	}
}

func (AWSDynamoDBOperations) injectSpanToCarrier(interface{}, ot.Span) error {
	return nil
}
