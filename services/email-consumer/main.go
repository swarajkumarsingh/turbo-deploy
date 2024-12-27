package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/conf"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/ses"
)

var sesSenderEmail string = "swaraj.singh.wearingo@gmail.com"

type Queue struct {
	AppName        string `json:"appName"`
	ProjectId      string `json:"projectId"`
	DeploymentId   string `json:"deploymentId"`
	RecipientEmail string `json:"recipient_email"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	Timestamp      string `json:"timestamp"`
}

func processMessage(sqsSvc *sqs.SQS, msg *sqs.Message) error {
	body := aws.StringValue(msg.Body)
	receiptHandle := msg.ReceiptHandle

	var queue Queue
	if err := json.Unmarshal([]byte(body), &queue); err != nil {
		log.Printf("Error un-marshalling message body: %v", err)
		return err
	}

	log.Printf("AppName: %s, Recipient Email: %s, ProjectId: %s, DeploymentId: %s, Timestamp: %s",
		queue.AppName, queue.RecipientEmail, queue.ProjectId, queue.DeploymentId, queue.Timestamp)

	// valid := utils.ValidEmail(queue.RecipientEmail)
	// if !valid {
	// 	log.Println("invalid email address: ", queue.RecipientEmail)
	// 	if err := deleteMessage(sqsSvc, receiptHandle); err != nil {
	// 		return err
	// 	}
	// }

	_, err := ses.SendEmail(sesSenderEmail, queue.RecipientEmail, queue.Subject, queue.Body, "test", "UTF-8")
	if err != nil {
		log.Printf("Error sending email: %v", err)
	}

	if err := deleteMessage(sqsSvc, receiptHandle); err != nil {
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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(conf.AWS_REGION),
	})
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}

	sqsSvc := sqs.New(sess)

	log.Println("Listening to messages")
	for {
		// Receive messages from SQS queue
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

		// Process received messages
		for _, msg := range msgResult.Messages {
			if err := processMessage(sqsSvc, msg); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}
