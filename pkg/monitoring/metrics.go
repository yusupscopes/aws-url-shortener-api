package monitoring

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/yusupscopes/aws-url-shortener-api/pkg/logger"
)

const (
	// Namespace for CloudWatch metrics
	Namespace = "URLShortener"
)

// Metric names
const (
	MetricURLCreated        = "URLCreated"
	MetricURLRedirected     = "URLRedirected"
	MetricURLNotFound       = "URLNotFound"
	MetricURLStatsRetrieved = "URLStatsRetrieved"
	MetricDynamoDBError     = "DynamoDBError"
	MetricAPILatency        = "APILatency"
)

// Dimensions
const (
	DimensionOperation = "Operation"
	DimensionEndpoint  = "Endpoint"
)

// Client is a wrapper for CloudWatch client
type Client struct {
	cwClient *cloudwatch.Client
}

// NewClient creates a new CloudWatch metrics client
func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load AWS config for CloudWatch", err)
		return nil, err
	}
	
	return &Client{
		cwClient: cloudwatch.NewFromConfig(cfg),
	}, nil
}

// PutMetric puts a metric data point to CloudWatch
func (c *Client) PutMetric(ctx context.Context, metricName string, value float64, dimensions ...types.Dimension) error {
	_, err := c.cwClient.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(Namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(value),
				Dimensions: dimensions,
				Timestamp:  aws.Time(time.Now()),
				Unit:       types.StandardUnitCount,
			},
		},
	})
	
	if err != nil {
		logger.Error("Failed to put metric data", map[string]interface{}{
			"metricName": metricName,
			"value":      value,
			"error":      err.Error(),
		})
		return err
	}
	
	logger.Debug("Successfully put metric data", map[string]interface{}{
		"metricName": metricName,
		"value":      value,
	})
	return nil
}

// RecordURLCreated records a URL creation event
func (c *Client) RecordURLCreated(ctx context.Context) error {
	return c.PutMetric(ctx, MetricURLCreated, 1.0, types.Dimension{
		Name:  aws.String(DimensionOperation),
		Value: aws.String("CreateURL"),
	})
}

// RecordURLRedirected records a URL redirection event
func (c *Client) RecordURLRedirected(ctx context.Context) error {
	return c.PutMetric(ctx, MetricURLRedirected, 1.0, types.Dimension{
		Name:  aws.String(DimensionOperation),
		Value: aws.String("RedirectURL"),
	})
}

// RecordURLNotFound records a URL not found event
func (c *Client) RecordURLNotFound(ctx context.Context) error {
	return c.PutMetric(ctx, MetricURLNotFound, 1.0, types.Dimension{
		Name:  aws.String(DimensionOperation),
		Value: aws.String("LookupURL"),
	})
}

// RecordURLStatsRetrieved records a URL stats retrieval event
func (c *Client) RecordURLStatsRetrieved(ctx context.Context) error {
	return c.PutMetric(ctx, MetricURLStatsRetrieved, 1.0, types.Dimension{
		Name:  aws.String(DimensionOperation),
		Value: aws.String("GetURLStats"),
	})
}

// RecordDynamoDBError records a DynamoDB error event
func (c *Client) RecordDynamoDBError(ctx context.Context, operation string) error {
	return c.PutMetric(ctx, MetricDynamoDBError, 1.0, types.Dimension{
		Name:  aws.String(DimensionOperation),
		Value: aws.String(operation),
	})
}

// RecordAPILatency records API latency
func (c *Client) RecordAPILatency(ctx context.Context, endpoint string, latencyMs float64) error {
	return c.PutMetric(ctx, MetricAPILatency, latencyMs, 
		types.Dimension{
			Name:  aws.String(DimensionEndpoint),
			Value: aws.String(endpoint),
		},
		types.Dimension{
			Name:  aws.String("Unit"),
			Value: aws.String("Milliseconds"),
		},
	)
}