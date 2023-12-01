// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
	"github.com/stretchr/testify/assert"
)

const (
	requestID = "aws-test-request-id"
	region    = "aws-test-region"
)

func TestDynamoDBGetObjectWithError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ps := sensor.Tracer().StartSpan("aws-parent-dynamodb-span")

	ctx := instana.ContextWithSpan(context.TODO(), ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	ddClient := dynamodb.NewFromConfig(cfg)
	bucket := "dynamodb-test-bucket"
	tableName := "dynamodb-test-table"
	movie := Movie{Title: bucket, Year: 878}

	_, err = ddClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       movie.GetKey(),
	})

	assert.NoError(t, err)

	ps.Finish()

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recordedSpans))

	ddSpan := recordedSpans[0]
	assert.IsType(t, instana.AWSDynamoDBSpanData{}, ddSpan.Data)

	data := ddSpan.Data.(instana.AWSDynamoDBSpanData)
	assert.Equal(t, instana.AWSDynamoDBSpanTags{
		Operation: "get",
		Table:     tableName,
		Region:    region,
	}, data.Tags)
}

func TestDynamoDBMonitoredOperations(t *testing.T) {
	//bucket := "dynamodb-test-bucket"
	//key := "dynamodb-test-key"
	tableName := "dynamodb-test-table"
	//movie := Movie{Title: "dynamodb-test-bucket", Year: 878}

	testcases := map[string]struct {
		MonitoredFunc func(ctx context.Context, client *dynamodb.Client) (interface{}, error)
		ExpectedOut   instana.AWSDynamoDBSpanTags
	}{
		"CreateTable": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.CreateTableInput{
					TableName:            &tableName,
					AttributeDefinitions: make([]types.AttributeDefinition, 0),
					KeySchema:            make([]types.KeySchemaElement, 0),
				}

				return client.CreateTable(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "create",
				Table:     tableName,
				Region:    region,
			},
		},
		"ListTables": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.ListTablesInput{
					ExclusiveStartTableName: &tableName,
				}

				return client.ListTables(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "list",
				Region:    region,
			},
		},
		"PutItem": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.PutItemInput{
					Item:      make(map[string]types.AttributeValue),
					TableName: &tableName,
				}

				return client.PutItem(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "put",
				Table:     tableName,
				Region:    region,
			},
		},
		"UpdateItem": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.UpdateItemInput{
					Key:       make(map[string]types.AttributeValue),
					TableName: &tableName,
				}

				return client.UpdateItem(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "update",
				Table:     tableName,
				Region:    region,
			},
		},
		"DeleteItem": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.DeleteItemInput{
					Key:       make(map[string]types.AttributeValue),
					TableName: &tableName,
				}

				return client.DeleteItem(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "delete",
				Table:     tableName,
				Region:    region,
			},
		},
		"Query": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.QueryInput{
					TableName: &tableName,
				}

				return client.Query(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "query",
				Table:     tableName,
				Region:    region,
			},
		},
		"Scan": {
			MonitoredFunc: func(ctx context.Context, client *dynamodb.Client) (interface{}, error) {
				ip := dynamodb.ScanInput{
					TableName: &tableName,
				}

				return client.Scan(ctx, &ip)
			},
			ExpectedOut: instana.AWSDynamoDBSpanTags{
				Operation: "scan",
				Table:     tableName,
				Region:    region,
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
			)

			ps := sensor.Tracer().StartSpan("aws-parent-dynamodb-span")

			ctx := instana.ContextWithSpan(context.TODO(), ps)

			cfg, err := config.LoadDefaultConfig(ctx)
			assert.NoError(t, err)

			cfg = applyTestingChanges(cfg)

			instaawsv2.Instrument(sensor, &cfg)

			ddClient := dynamodb.NewFromConfig(cfg)

			_, err = testcase.MonitoredFunc(ctx, ddClient)

			assert.NoError(t, err)

			ps.Finish()

			recordedSpans := recorder.GetQueuedSpans()
			assert.Equal(t, 2, len(recordedSpans))

			ddSpan := recordedSpans[0]
			assert.IsType(t, instana.AWSDynamoDBSpanData{}, ddSpan.Data)

			data := ddSpan.Data.(instana.AWSDynamoDBSpanData)
			assert.Equal(t, testcase.ExpectedOut, data.Tags)
		})
	}

}

func TestDynamoDBNoParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	ddClient := dynamodb.NewFromConfig(cfg)
	bucket := "dynamodb-test-bucket"
	tableName := "dynamodb-test-table"
	movie := Movie{Title: bucket, Year: 878}

	_, err = ddClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       movie.GetKey(),
	})

	assert.Error(t, err) //error is fine as we are more interested in the span details. Mocking the response data should solve this.

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recordedSpans))
}

func TestDynamoDBUnMonitoredMethod(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(sensor, &cfg)

	ddClient := dynamodb.NewFromConfig(cfg)
	backupName := "dynamodb-test-backup"
	tableName := "dynamodb-test-table"

	_, err = ddClient.CreateBackup(ctx, &dynamodb.CreateBackupInput{
		BackupName: &backupName,
		TableName:  &tableName,
	})

	assert.Error(t, err) //error is fine as we are more interested in the span details. Mocking the response data should solve this.

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recordedSpans))
}

type Movie struct {
	Title string                 `dynamodbav:"title"`
	Year  int                    `dynamodbav:"year"`
	Info  map[string]interface{} `dynamodbav:"info"`
}

// GetKey returns the composite primary key of the movie in a format that can be
// sent to DynamoDB.
func (movie Movie) GetKey() map[string]types.AttributeValue {
	title, err := attributevalue.Marshal(movie.Title)
	if err != nil {
		panic(err)
	}
	year, err := attributevalue.Marshal(movie.Year)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"title": title, "year": year}
}
