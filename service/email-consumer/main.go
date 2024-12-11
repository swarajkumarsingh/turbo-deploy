package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/conf"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/utils"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/ses"
)

const (
	sesSenderEmail = "swaraj.singh.wearingo@gmail.com"
	sesSampleRecipientEmail = "swaraj.singh.wearingo@gmail.com"
)

func main() {
	// Initialize AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(conf.AWS_REGION),
	})
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}

	// Create SQS and SES service clients
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
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		// Process received messages
		for _, msg := range msgResult.Messages {

			recipientEmail := aws.StringValue(msg.Body)
			fmt.Printf("Sending email to %s\n", recipientEmail)

			// validate email
			valid := utils.ValidEmail(recipientEmail)
			if !valid {
				log.Println("invalid email address: ", recipientEmail)
				continue
			}
			
			_, err := ses.SendEmail(sesSenderEmail, sesSampleRecipientEmail, "subject", "<h1>test</h1>", "test", "UTF-8")
			if err != nil {
				log.Printf("Error sending email: %v", err)
			}

			// Delete the processed message from the SQS queue
			_, err = sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(conf.AWS_SQS_URL),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Println("Error deleting message:", err)
				continue
			}
		}

		// Wait before checking for new messages again
		time.Sleep(500 * time.Millisecond)
	}
}
