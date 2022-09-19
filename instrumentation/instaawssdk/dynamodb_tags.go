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
	var tags opentracing.Tags
	switch params := req.Params.(type) {
	case *dynamodb.CreateTableInput:
		tags = opentracing.Tags{
			dynamodbOp:    "create",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	case *dynamodb.ListTablesInput:
		tags = opentracing.Tags{
			dynamodbOp: "list",
		}
	case *dynamodb.GetItemInput:
		tags = opentracing.Tags{
			dynamodbOp:    "get",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	case *dynamodb.PutItemInput:
		tags = opentracing.Tags{
			dynamodbOp:    "put",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	case *dynamodb.UpdateItemInput:
		tags = opentracing.Tags{
			dynamodbOp:    "update",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	case *dynamodb.DeleteItemInput:
		tags = opentracing.Tags{
			dynamodbOp:    "delete",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	case *dynamodb.QueryInput:
		tags = opentracing.Tags{
			dynamodbOp:    "query",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	case *dynamodb.ScanInput:
		tags = opentracing.Tags{
			dynamodbOp:    "scan",
			dynamodbTable: aws.StringValue(params.TableName),
		}
	default:
		return nil, errMethodNotInstrumented
	}

	tags[dynamodbRegion] = aws.StringValue(req.Config.Region)

	return tags, nil
}
