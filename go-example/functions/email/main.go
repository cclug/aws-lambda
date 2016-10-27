package main

import (
    "fmt"
    "bytes"
    "io/ioutil"
//    "net/mail"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    
    "github.com/veqryn/go-email/email"
)

func main() {
  region := "us-east-1"
  sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
  if err != nil {
    fmt.Println("failed to create session,", err)
    return
  }
  
  svc := s3.New(sess)
  params := &s3.GetObjectInput {
    Bucket:                     aws.String("devspire-ses"), // Required
    Key:                        aws.String("pvnu7delckle60vnrmltf7nut04mrpisdqn3o0g1"),  // Required
  }
  

  resp, err := svc.GetObject(params)

  if err != nil {
    // Print the error, cast err to awserr.Error to get the Code and
    // Message from an error.
    fmt.Println(err.Error())
    return
   }

  // Pretty-print the response data.
  fmt.Println(resp)
  
//  b, err := ioutil.ReadAll(resp.Body)
//  if err != nil {
//     fmt.Println(err)
//  }
  
  //m, err := mail.ReadMessage(resp.Body)
  //body, err := ioutil.ReadAll(m.Body)
  fullBody, err := ioutil.ReadAll(resp.Body)
    
//  fmt.Printf("%s", body)

  reader := bytes.NewReader(fullBody)
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
