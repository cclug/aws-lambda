package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/veqryn/go-email/email"
)

// make these variables so we can change them when testing
// eventually we need to get these programmatically
var region = "us-east-1"
var bucket = "devspire-ses"
var address = "xxxx@gmail.com"
var subject = "test subject"
var messageId = "need to get this from event"

func main() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}
	body, err := getBody(sess, bucket, messageId)
	if err != nil {
		log.Fatal(err)
	}
	text, err := getText(body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", text)
	err = sendEmail(sess, address, address, text, subject)
	if err != nil {
		log.Fatal(err)
	}
}

// sendEmail: send email (and print response to stderr)
func sendEmail(sess *session.Session, to, from, text, subject string) error {
	svc := ses.New(sess)
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#example_SES_SendEmail
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{ // Required
			ToAddresses: []*string{
				aws.String(to), // Required
			},
		},
		Message: &ses.Message{ // Required
			Body: &ses.Body{ // Required
				Text: &ses.Content{
					Data:    aws.String(text), // Required
					Charset: aws.String("UTF-8"),
				},
			},
			Subject: &ses.Content{ // Required
				Data: aws.String(subject), // Required
			},
		},
		Source: aws.String(from), // Required
	}
	resp, err := svc.SendEmail(params)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v\n", resp)
	return nil
}

// getText: get text from body of email
func getText(body []byte) (string, error) {
	reader := bytes.NewReader(body)
	msg, err := email.ParseMessage(reader)
	if err != nil {
		return "", err
	}

	var text string
	for _, part := range msg.PartsContentTypePrefix("text/plain") {
		text = string(part.Body)
		// fmt.Println(part.Header["Content-Type"])
		// todo: parse "[text/plain; charset=UTF-8]"
	}
	return text, nil
}

// getBody: get email body from s3 bucket
func getBody(sess *session.Session, bucket, key string) ([]byte, error) {
	svc := s3.New(sess)
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket), // required
		Key:    aws.String(key),    // required
	}
	resp, err := svc.GetObject(params)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}
