package ses

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/swarajkumarsingh/go-email-consumer-aws-sqs/conf"
)

func SendEmail(sender, recipient, subject, htmlBody, textBody, charSet string) (*ses.SendEmailOutput, error) {
	var result *ses.SendEmailOutput

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(conf.AWS_REGION),
			Credentials: credentials.NewStaticCredentials(conf.AWS_ACCESS_KEY, conf.AWS_SECRET_ACCESS_KEY, conf.AWS_TOKEN),
		},
	})

	if err != nil {
		log.Println(err)
		return result, errors.New("something went wrong")
	}

	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	result, err = svc.SendEmail(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
			case ses.ErrCodeMailFromDomainNotVerifiedException:
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
				return result, errors.New("something went wrong")
			default:
				log.Println(aerr.Error())
			}
		}
		log.Println(err)
		return result, errors.New("something went wrong")
	}
	log.Println("Email Sent to address: " + recipient)
	return result, err
}
