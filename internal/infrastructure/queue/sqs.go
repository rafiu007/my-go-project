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
    
    "my_go_project/internal/domain/entity"
    "my_go_project/internal/infrastructure/db"
)

type SQSClient struct {
    client   *sqs.Client
    queueURL string
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

func (s *SQSClient) StartConsumer(ctx context.Context) {
    go func() {
        log.Println("Starting SQS consumer...")
        for {
            select {
            case <-ctx.Done():
                log.Println("Stopping SQS consumer...")
                return
            default:
                s.consumeMessages(ctx)
                time.Sleep(time.Second * 5)
            }
        }
    }()
}

func (s *SQSClient) consumeMessages(ctx context.Context) {
    input := &sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(s.queueURL),
        MaxNumberOfMessages: 10,
        WaitTimeSeconds:     5,
    }

    output, err := s.client.ReceiveMessage(ctx, input)
    if err != nil {
        log.Printf("Error receiving messages: %v", err)
        return
    }

    for _, msg := range output.Messages {
        var entry entity.CalendarEntry
        if err := json.Unmarshal([]byte(*msg.Body), &entry); err != nil {
            log.Printf("Error unmarshaling message: %v", err)
            continue
        }

        fmt.Printf("\nReceived Calendar Entry:\n")
        fmt.Printf("ID: %d\n", entry.ID)
        fmt.Printf("Start Date: %s\n", entry.StartDate.Format("2006-01-02 15:04:05"))
        fmt.Printf("Stop Date: %s\n", entry.StopDate.Format("2006-01-02 15:04:05"))
        fmt.Printf("Created At: %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))

        deleteInput := &sqs.DeleteMessageInput{
            QueueUrl:      aws.String(s.queueURL),
            ReceiptHandle: msg.ReceiptHandle,
        }
        
        if _, err := s.client.DeleteMessage(ctx, deleteInput); err != nil {
            log.Printf("Error deleting message: %v", err)
        }
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