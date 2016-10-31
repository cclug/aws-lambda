package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/apex/go-apex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/veqryn/go-email/email"
)

const (
	bucket = "devspire-ses"
	region = "us-east-1" // see handle()
)

// must be all lower case
var whitelist = []string{
	"mkuchin@gmail.com",            // Max
	"me@tobin.cc",                  // Tobin
	"tstarling@wikimedia.org",      // Tim
	"neville.bagot@det.nsw.edu.au", // Neville
	"robert@thorsby.com.au",        // Robert
	"officemail2259@yahoo.com.au",  // Toby
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		if err := handle(event); err != nil {
			return "", err
		}
		return "", nil // do we need to return anything on success?
	})
}

// data structure used to unmarshal event
type record struct {
	ses struct {
		mail struct {
			messageId     string
			commonHeaders struct {
				returnPath string
				from       []string
				date       string
				to         []string
				messageId  string
				subject    string
			}
		}
	}
}

// handle an event
func handle(event json.RawMessage) error {
	r, err := getRecord(event)
	if err != nil {
		return err
	}
	// set region in ~/.aws/config or call
	// session.NewSession(&aws.Config{Region: aws.String(region)})
	sess := session.Must(session.NewSession())
	body, err := getBody(sess, bucket, r.ses.mail.messageId)
	if err != nil {
		return err
	}
	text, err := getText(body)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s\n", text)

	from := r.ses.mail.commonHeaders.from[0] // only accept single sender
	if !isAuthSender(from) {
		return fmt.Errorf("sender is not in whitelist: %s", from)
	}
	err = sendEmail(sess,
		from,
		text,
		r.ses.mail.commonHeaders.subject)
	if err != nil {
		return err
	}
	return nil
}

// getRecord: unmarshal record from event
func getRecord(event json.RawMessage) (record, error) {
	var records []record
	if err := json.Unmarshal(event, &records); err != nil {
		return record{}, err
	}
	if len(records) > 1 {
		fmt.Fprintf(os.Stderr, "multiple records unexpected: %d\n", len(records))
	}
	return records[0], nil
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

// isAuthSender: checks if address is authorized to send email
func isAuthSender(s string) bool {
	// address could be form "John Smith <john@domain.com>" or "john@domain.com"

	var validEmailAddress = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	addr := validEmailAddress.FindString(s)
	addr = strings.ToLower(addr)
	if addr == "" {
		return false
	}
	for _, valid := range whitelist {
		if addr == valid {
			return true
		}
	}
	return false
}

// sendEmail: send email (and print response to stderr)
func sendEmail(sess *session.Session, from, text, subject string) error {
	to := whitelistPtrs()
	svc := ses.New(sess)
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#example_SES_SendEmail
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{ // Required
			ToAddresses: to,
		},
		Message: &ses.Message{ // Required
			Body: &ses.Body{ // Required
				Text: &ses.Content{
					Data: aws.String(text), // Required
					// Charset: aws.String("UTF-8"),
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

// return slice of pointers to whitelisted email addresses
// used to initialize ses.SendEmailInput data structure
func whitelistPtrs() []*string {
	whitelist := []string{"mkuchin@gmailcom", "me@tobin.cc"} // remove this for production
	to := make([]*string, len(whitelist))
	for i := range whitelist {
		to[i] = &whitelist[i]
	}
	return to
}
