package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/veqryn/go-email/email"
)

func main() {

        bucket := "devspire-ses"
        messageId := "pvnu7delckle60vnrmltf7nut04mrpisdqn3o0g1"
        
        body, err := getBody(bucket, messageId)
        if err != nil {
                fmt.Println(err.Error())
                return
        }

	reader := bytes.NewReader(body)
	msg, err := email.ParseMessage(reader)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	for _, part := range msg.PartsContentTypePrefix("text/plain") {
		fmt.Println(string(part.Body))
	}

}

func getBody(bucket, key string) ([]byte, error) {
	region := "us-east-1"
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
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
