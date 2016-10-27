package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/veqryn/go-email/email"
)

var connection *session.Session

func main() {

	region := "us-east-1"
	bucket := "devspire-ses"

        var err error
	connection, err = session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}



//	messageId := "pvnu7delckle60vnrmltf7nut04mrpisdqn3o0g1"

	text, err := getTextBody(bucket, messageId)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
        err = sendEmail(text, "test subj")
        if err != nil {
                log.Fatal(err)
        } 

}

func sendEmail(text, subject string) error {
        email := "xxxx@gmail.com"
	svc := ses.New(connection)

        // https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#example_SES_SendEmail

	params := &ses.SendEmailInput{
		Destination: &ses.Destination{ // Required			
			ToAddresses: []*string{
				aws.String(email), // Required
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
				Data:    aws.String(subject), // Required
			},
		},
		Source: aws.String(email), // Required
	}
	resp, err := svc.SendEmail(params)

	if err != nil {
		return err
	}

	// Pretty-print the response data.
	fmt.Println(resp)
        return nil
}

func getTextBody(bucket, key string) (string, error) {
	body, err := getBody(bucket, key)
	if err != nil {
		return "", err
	}

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

func getBody(bucket, key string) ([]byte, error) {
	svc := s3.New(connection)
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket), // required
		Key:    aws.String(key),    // required
	}

	resp, err := svc.GetObject(params)
	if err != nil {
		return nil, err
	}

	fullBody, err := ioutil.ReadAll(resp.Body)
	return fullBody, err
}
