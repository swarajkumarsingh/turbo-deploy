package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
	"github.com/swarajkumarsingh/status-sqs-consumer/conf"
	"github.com/swarajkumarsingh/status-sqs-consumer/db"
)

var database = db.Mgr.DBConn

type Queue struct {
	AppName      string `json:"appName"`
	ProjectId    string `json:"projectId"`
	DeploymentId string `json:"deploymentId"`
	Host         string `json:"host"`
	Status       string `json:"Status"`
	Timestamp    string `json:"timestamp"`
}

func processMessage(sqsSvc *sqs.SQS, msg *sqs.Message) error {
	body := aws.StringValue(msg.Body)
	receiptHandle := msg.ReceiptHandle

	var queue Queue
	if err := json.Unmarshal([]byte(body), &queue); err != nil {
		log.Printf("Error un-marshalling message body: %v", err)
		return err
	}

	log.Printf("AppName: %s, ProjectId: %s, DeploymentId: %s, Host: %s, Status: %s, Timestamp: %s",
		queue.AppName, queue.ProjectId, queue.DeploymentId, queue.Host, queue.Status, queue.Timestamp)

	if queue.Status == "" {
		log.Println("Deleting message with empty 'Message' field")
		if err := deleteMessage(sqsSvc, receiptHandle); err != nil {
			return err
		}
		return nil
	}

	if err := CreateDeploymentLog(context.Background(), queue); err != nil {
		log.Println("Error while pushing to DB:", err.Error())
		return err
	}

	if err := deleteMessage(sqsSvc, receiptHandle); err != nil {
		return err
	}

	return nil
}

func CreateDeploymentLog(context context.Context, body Queue) error {
	if body.Status == "" {
		body.Status = "PROG"
	}
	query := `UPDATE deployments SET status = $1, updated_at = NOW() WHERE id = $2`

	_, err := database.ExecContext(context, query, body.Status, body.DeploymentId)
	if err != nil {
		return err
	}
	return nil
}

func deleteMessage(sqsSvc *sqs.SQS, receiptHandle *string) error {
	_, err := sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(conf.AWS_SQS_URL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		log.Printf("Error deleting message: %v", err)
	}
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(conf.AWS_REGION),
	})
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}

	sqsSvc := sqs.New(sess)
	log.Println("Listening to messages")

	for {
		msgResult, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(conf.AWS_SQS_URL),
			MaxNumberOfMessages: aws.Int64(conf.MaxNumberOfMessages),
			VisibilityTimeout:   aws.Int64(conf.VisibilityTimeout),
		})
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range msgResult.Messages {
			if err := processMessage(sqsSvc, msg); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}
