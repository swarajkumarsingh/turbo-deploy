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
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/conf"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/db"
)

var database = db.Mgr.DBConn

type Queue struct {
	AppName      string `json:"appName"`
	Message      string `json:"message"`
	LogType      string `json:"logType"`
	ProjectId    string `json:"projectId"`
	DeploymentId string `json:"deploymentId"`
	Environment  string `json:"environment"`
	Host         string `json:"host"`
	Cause        string `json:"cause"`
	Name         string `json:"name"`
	Stack        string `json:"stack"`
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

	log.Printf("AppName: %s, Message: %s, LogType: %s, ProjectId: %s, DeploymentId: %s, Host: %s, Environment: %s, Cause: %s, Stack: %s, Timestamp: %s",
		queue.AppName, queue.Message, queue.LogType, queue.ProjectId, queue.DeploymentId, queue.Host, queue.Environment, queue.Cause, queue.Stack, queue.Timestamp)

	if queue.Message == "" {
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
	if body.LogType == "" {
		body.LogType = "INFO"
	}
	query := `INSERT INTO deployment_logs(deployment_id, project_id, environment, message, cause, stack, name, host, log_type, timestamp) 
          VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := database.ExecContext(context, query, body.DeploymentId, body.ProjectId, body.Environment, body.Message, body.Cause, body.Stack, body.Name, body.Host, body.LogType, body.Timestamp)
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
