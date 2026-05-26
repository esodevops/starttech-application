package logger

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchWriter implements io.Writer to send logs to AWS CloudWatch Logs
// This is a minimal example; production code should handle batching, errors, and sequence tokens robustly.
type CloudWatchWriter struct {
	logGroupName  string
	logStreamName string
	client        *cloudwatchlogs.Client
	sequenceToken *string
}

func NewCloudWatchWriter(logGroup, logStream string) (*CloudWatchWriter, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := cloudwatchlogs.NewFromConfig(cfg)
	return &CloudWatchWriter{
		logGroupName:  logGroup,
		logStreamName: logStream,
		client:        client,
	}, nil
}

func (w *CloudWatchWriter) Write(p []byte) (n int, err error) {
	ctx := context.Background()
	       input := &cloudwatchlogs.PutLogEventsInput{
		       LogEvents: []types.InputLogEvent{
			       {
				       Message:   aws.String(string(p)),
				       Timestamp: aws.Int64(time.Now().UnixMilli()),
			       },
		       },
		       LogGroupName:  aws.String(w.logGroupName),
		       LogStreamName: aws.String(w.logStreamName),
	       }
	if w.sequenceToken != nil {
		input.SequenceToken = w.sequenceToken
	}
	resp, err := w.client.PutLogEvents(ctx, input)
	if err != nil {
		return 0, err
	}
	w.sequenceToken = resp.NextSequenceToken
	return len(p), nil
}
