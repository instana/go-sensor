package instana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpanCategoryString(t *testing.T) {
	tests := []struct {
		name     string
		category spanCategory
		expected string
	}{
		{
			name:     "HTTP category",
			category: httpReq,
			expected: "http",
		},
		{
			name:     "RPC category",
			category: rpc,
			expected: "rpc",
		},
		{
			name:     "Messaging category",
			category: messaging,
			expected: "messaging",
		},
		{
			name:     "Logging category",
			category: logging,
			expected: "logging",
		},
		{
			name:     "Databases category",
			category: databases,
			expected: "databases",
		},
		{
			name:     "GraphQL category",
			category: graphql,
			expected: "graphql",
		},
		{
			name:     "Unknown category",
			category: unknown,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.category.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTracerOptionsDisableAllCategories(t *testing.T) {
	opts := &TracerOptions{}
	opts.DisableAllCategories()

	expectedCategories := []spanCategory{httpReq, rpc, messaging, logging, databases, graphql}

	// Check if all categories are disabled
	for _, category := range expectedCategories {
		if !opts.Disable[category.String()] {
			t.Errorf("Category %s should be disabled", category)
		}
	}

	// Check if the map has the correct size
	if len(opts.Disable) != len(expectedCategories) {
		t.Errorf("Expected %d disabled categories, got %d", len(expectedCategories), len(opts.Disable))
	}
}

func TestSpanSGetSpanCategory(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		expectedCat spanCategory
	}{
		{
			name:        "HTTP Server span",
			operation:   string(HTTPServerSpanType),
			expectedCat: httpReq,
		},
		{
			name:        "HTTP Client span",
			operation:   string(HTTPClientSpanType),
			expectedCat: httpReq,
		},
		{
			name:        "RPC Server span",
			operation:   string(RPCServerSpanType),
			expectedCat: rpc,
		},
		{
			name:        "RPC Client span",
			operation:   string(RPCClientSpanType),
			expectedCat: rpc,
		},
		{
			name:        "Kafka span",
			operation:   string(KafkaSpanType),
			expectedCat: messaging,
		},
		{
			name:        "RabbitMQ span",
			operation:   string(RabbitMQSpanType),
			expectedCat: messaging,
		},
		{
			name:        "Log span",
			operation:   string(LogSpanType),
			expectedCat: logging,
		},
		{
			name:        "DynamoDB span",
			operation:   string(AWSDynamoDBSpanType),
			expectedCat: databases,
		},
		{
			name:        "S3 span",
			operation:   string(AWSS3SpanType),
			expectedCat: databases,
		},
		{
			name:        "MongoDB span",
			operation:   string(MongoDBSpanType),
			expectedCat: databases,
		},
		{
			name:        "PostgreSQL span",
			operation:   string(PostgreSQLSpanType),
			expectedCat: databases,
		},
		{
			name:        "MySQL span",
			operation:   string(MySQLSpanType),
			expectedCat: databases,
		},
		{
			name:        "Redis span",
			operation:   string(RedisSpanType),
			expectedCat: databases,
		},
		{
			name:        "Couchbase span",
			operation:   string(CouchbaseSpanType),
			expectedCat: databases,
		},
		{
			name:        "Cosmos span",
			operation:   string(CosmosSpanType),
			expectedCat: databases,
		},
		{
			name:        "GraphQL Server span",
			operation:   string(GraphQLServerType),
			expectedCat: graphql,
		},
		{
			name:        "GraphQL Client span",
			operation:   string(GraphQLClientType),
			expectedCat: graphql,
		},
		{
			name:        "Unknown span type",
			operation:   "unknown-span-type",
			expectedCat: unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := &spanS{
				Operation: tt.operation,
			}

			result := span.getSpanCategory()

			if result != tt.expectedCat {
				t.Errorf("Expected category %s, got %s", tt.expectedCat, result)
			}
		})
	}
}

func TestSpanCategoryEnabled(t *testing.T) {
	tests := []struct {
		name     string
		category spanCategory
		disable  map[string]bool
		expected bool
	}{
		{
			name:     "Category enabled when no categories are disabled",
			category: httpReq,
			disable:  map[string]bool{},
			expected: true,
		},
		{
			name:     "Category disabled when specifically disabled",
			category: httpReq,
			disable:  map[string]bool{"http": true},
			expected: false,
		},
		{
			name:     "Category enabled when other categories are disabled",
			category: httpReq,
			disable:  map[string]bool{"rpc": true, "messaging": true},
			expected: true,
		},
		{
			name:     "Category disabled when all categories are disabled",
			category: databases,
			disable: map[string]bool{
				"http":      true,
				"rpc":       true,
				"messaging": true,
				"logging":   true,
				"databases": true,
				"graphql":   true,
			},
			expected: false,
		},
		{
			name:     "GraphQL category disabled when specifically disabled",
			category: graphql,
			disable:  map[string]bool{"graphql": true},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			InitCollector(&Options{
				Tracer: TracerOptions{
					Disable: tc.disable,
				},
			})
			defer ShutdownCollector()

			result := tc.category.Enabled()

			assert.Equal(t, tc.expected, result)
		})
	}
}
