// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instaawssdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/opentracing/opentracing-go"
)

func extractDynamoDBTags(req *request.Request) (opentracing.Tags, error) {
	switch params := req.Params.(type) {
	case *dynamodb.CreateTableInput:
		return opentracing.Tags{
			"dynamodb.op":    "create",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.ListTablesInput:
		return opentracing.Tags{
			"dynamodb.op": "list",
		}, nil
	case *dynamodb.GetItemInput:
		return opentracing.Tags{
			"dynamodb.op":    "get",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.PutItemInput:
		return opentracing.Tags{
			"dynamodb.op":    "put",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.UpdateItemInput:
		return opentracing.Tags{
			"dynamodb.op":    "update",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.DeleteItemInput:
		return opentracing.Tags{
			"dynamodb.op":    "delete",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.QueryInput:
		return opentracing.Tags{
			"dynamodb.op":    "query",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.ScanInput:
		return opentracing.Tags{
			"dynamodb.op":    "scan",
			"dynamodb.table": aws.StringValue(params.TableName),
		}, nil
	default:
		return nil, errMethodNotInstrumented
	}
}
