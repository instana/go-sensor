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
			dynamodbOp:    "create",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.ListTablesInput:
		return opentracing.Tags{
			dynamodbOp: "list",
		}, nil
	case *dynamodb.GetItemInput:
		return opentracing.Tags{
			dynamodbOp:    "get",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.PutItemInput:
		return opentracing.Tags{
			dynamodbOp:    "put",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.UpdateItemInput:
		return opentracing.Tags{
			dynamodbOp:    "update",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.DeleteItemInput:
		return opentracing.Tags{
			dynamodbOp:    "delete",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.QueryInput:
		return opentracing.Tags{
			dynamodbOp:    "query",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	case *dynamodb.ScanInput:
		return opentracing.Tags{
			dynamodbOp:    "scan",
			dynamodbTable: aws.StringValue(params.TableName),
		}, nil
	default:
		return nil, errMethodNotInstrumented
	}
}
