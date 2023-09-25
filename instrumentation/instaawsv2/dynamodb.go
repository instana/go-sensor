// (c) Copyright IBM Corp. 2023

package instaawsv2

import (
	"context"
	"errors"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var errUnknownDynamoDBMethod = errors.New("dynamodb method not instrumented")

func injectAWSContextWithDynamoDBSpan(tr instana.TracerLogger, ctx context.Context, params interface{}) context.Context {
	tags, err := extractDynamoDBTags(params)
	if err != nil {
		if errors.Is(err, errUnknownDynamoDBMethod) {
			tr.Logger().Error("failed to identify the dynamodb method: ", err.Error())
			return ctx
		}
	}

	// By design, we will abort the dynamodb span creation if a parent span is not identified.
	parent, ok := instana.SpanFromContext(ctx)
	if !ok {
		tr.Logger().Error("failed to retrieve the parent span. Aborting dynamodb child span creation.")
		return ctx
	}

	sp := tr.Tracer().StartSpan("dynamodb",
		ext.SpanKindRPCClient,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{
			dynamodbRegion: awsmiddleware.GetRegion(ctx),
		},
		tags,
	)

	return instana.ContextWithSpan(ctx, sp)
}

// finishDynamoDBSpan retrieves tags from completed calls and adds them to the span
func finishDynamoDBSpan(tr instana.TracerLogger, ctx context.Context, err error) {
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

func extractDynamoDBTags(params interface{}) (opentracing.Tags, error) {
	switch params := params.(type) {
	case *dynamodb.CreateTableInput:
		return opentracing.Tags{
			dynamodbOp:    "create",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.ListTablesInput:
		return opentracing.Tags{
			dynamodbOp: "list",
		}, nil
	case *dynamodb.GetItemInput:
		return opentracing.Tags{
			dynamodbOp:    "get",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.PutItemInput:
		return opentracing.Tags{
			dynamodbOp:    "put",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.UpdateItemInput:
		return opentracing.Tags{
			dynamodbOp:    "update",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.DeleteItemInput:
		return opentracing.Tags{
			dynamodbOp:    "delete",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.QueryInput:
		return opentracing.Tags{
			dynamodbOp:    "query",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	case *dynamodb.ScanInput:
		return opentracing.Tags{
			dynamodbOp:    "scan",
			dynamodbTable: stringDeRef(params.TableName),
		}, nil
	default:
		return nil, errUnknownDynamoDBMethod
	}
}
