// internal/infrastructure/queue/sqs.go
package queue

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    awsconfig "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
    
    "my_go_project/internal/domain/entity"
    "my_go_project/internal/infrastructure/db"
)

type SQSClient struct {
    client   *sqs.Client
    queueURL string
	isConsuming bool        // Track consumer status
    stopChan    chan bool   // Channel for stopping consumer
	
}

func NewSQSClient(endpoint, queueURL string) (*SQSClient, error) {
    customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
        return aws.Endpoint{
            URL:           endpoint,
            SigningRegion: "us-east-1",
            HostnameImmutable: true,
        }, nil
    })

    cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
        awsconfig.WithRegion("us-east-1"),
        awsconfig.WithEndpointResolverWithOptions(customResolver),
        awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
    )
    if err != nil {
        return nil, fmt.Errorf("unable to load SDK config: %v", err)
    }

    client := sqs.NewFromConfig(cfg)
    
    log.Printf("Initialized SQS client with endpoint: %s and queue URL: %s", endpoint, queueURL)
    
    return &SQSClient{
        client:   client,
        queueURL: queueURL,
		isConsuming: false,
		stopChan: make(chan bool),
    }, nil
}

func (s *SQSClient) SendMessages(ctx context.Context, entries []entity.CalendarEntry) error {
    log.Printf("Starting to send %d entries to queue", len(entries))
    
    for _, entry := range entries {
        data, err := json.Marshal(entry)
        if err != nil {
            log.Printf("Error marshaling entry %d: %v", entry.ID, err)
            continue
        }

        input := &sqs.SendMessageInput{
            QueueUrl:    aws.String(s.queueURL),
            MessageBody: aws.String(string(data)),
        }

        _, err = s.client.SendMessage(ctx, input)
        if err != nil {
            log.Printf("Error sending message for entry %d: %v", entry.ID, err)
            continue
        }

        log.Printf("Successfully sent entry %d to queue", entry.ID)
    }
    return nil
}

// Improved StartConsumer with better error handling and shutdown
func (s *SQSClient) StartConsumer(ctx context.Context) {
    if s.isConsuming {
        log.Println("Consumer is already running")
        return
    }

    s.isConsuming = true
    go func() {
        log.Println("Starting SQS consumer...")
        defer func() {
            s.isConsuming = false
            log.Println("SQS consumer stopped")
        }()

        for {
            select {
            case <-ctx.Done():
                log.Println("Context cancelled, stopping consumer...")
                return
            case <-s.stopChan:
                log.Println("Received stop signal, stopping consumer...")
                return
            default:
                if err := s.consumeMessages(ctx); err != nil {
                    log.Printf("Error consuming messages: %v", err)
                    time.Sleep(time.Second * 5) // Back off on error
                }
            }
        }
    }()
}

// Update the consumeMessages function
func (s *SQSClient) consumeMessages(ctx context.Context) error {
    log.Println("Polling for messages...")
    
    input := &sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(s.queueURL),
        MaxNumberOfMessages: 10,
        WaitTimeSeconds:     5,
        VisibilityTimeout:   30,
        AttributeNames:      []types.QueueAttributeName{types.QueueAttributeNameAll},
        MessageAttributeNames: []string{"All"},
    }

    output, err := s.client.ReceiveMessage(ctx, input)
    if err != nil {
        return fmt.Errorf("error receiving messages: %v", err)
    }

    if len(output.Messages) > 0 {
        log.Printf("Processing %d messages", len(output.Messages))
    }

    for _, msg := range output.Messages {
        if err := s.processMessage(ctx, msg); err != nil {
            log.Printf("Error processing message %s: %v", *msg.MessageId, err)
            continue
        }
    }

    return nil
}

// Update the processMessage function signature
func (s *SQSClient) processMessage(ctx context.Context, msg types.Message) error {
    log.Printf("Processing message: %s", *msg.MessageId)

    // Handle test messages differently
    if *msg.Body == "test message" {
        log.Printf("Received test message: %s", *msg.Body)
        return s.deleteMessage(ctx, msg.ReceiptHandle)
    }

    var entry entity.CalendarEntry
    if err := json.Unmarshal([]byte(*msg.Body), &entry); err != nil {
        return fmt.Errorf("error unmarshaling message: %v", err)
    }

    // Print formatted entry
    fmt.Printf("\n=== Calendar Entry ===\n")
    fmt.Printf("Message ID: %s\n", *msg.MessageId)
    fmt.Printf("ID: %d\n", entry.ID)
    fmt.Printf("Start Date: %s\n", entry.StartDate.Format("2006-01-02 15:04:05"))
    fmt.Printf("Stop Date: %s\n", entry.StopDate.Format("2006-01-02 15:04:05"))
    fmt.Printf("Created At: %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))
    fmt.Printf("Updated At: %s\n", entry.UpdatedAt.Format("2006-01-02 15:04:05"))
    fmt.Printf("==================\n")

    // Delete message after successful processing
    if err := s.deleteMessage(ctx, msg.ReceiptHandle); err != nil {
        return fmt.Errorf("error deleting message: %v", err)
    }

    log.Printf("Successfully processed and deleted message: %s", *msg.MessageId)
    return nil
}

// Update deleteMessage function if needed
func (s *SQSClient) deleteMessage(ctx context.Context, receiptHandle *string) error {
    input := &sqs.DeleteMessageInput{
        QueueUrl:      aws.String(s.queueURL),
        ReceiptHandle: receiptHandle,
    }
    _, err := s.client.DeleteMessage(ctx, input)
    return err
}

// Method to stop the consumer
func (s *SQSClient) StopConsumer() {
    if s.isConsuming {
        s.stopChan <- true
    }
}

 

func (s *SQSClient) StartScheduler(ctx context.Context, database *db.Database) {
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for {
            select {
            case <-ctx.Done():
                ticker.Stop()
                return
            case <-ticker.C:
                var entries []entity.CalendarEntry
                if err := database.Where("stop_date > ?", time.Now()).Find(&entries).Error; err != nil {
                    log.Printf("Error fetching entries: %v", err)
                    continue
                }

                if err := s.SendMessages(ctx, entries); err != nil {
                    log.Printf("Error sending scheduled messages: %v", err)
                }
            }
        }
    }()
}