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
	"gopkg.in/yaml.v2"
)

const (
	nl = "\r\n"
)

type Config struct {
	Bucket     string `yaml:"bucket"`
	InboxEmail string `yaml:"inboxEmail"`
	Whitelist  []string `yaml:"whitelist"`
}

var config Config

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		if err := handle(event); err != nil {
			return "", err
		}
		return "", nil // do we need to return anything on success?
	})
}

// simplified mail event
type mail struct {
	messageId string
	headers   struct {
		from      []string
		messageId string
		subject   string
	}
}

// handle an event
func handle(event json.RawMessage) error {
	//        fmt.Println("Executing Lambda function")
	m, err := eventToMail(event)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return fmt.Errorf("Error reading config: %s", err.Error())
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("Error unmarshalling config: %s", err.Error())
	}

        fmt.Fprintf(os.Stderr, "Config: %v\n", config)

	sess := session.Must(session.NewSession())
	body, err := getBody(sess, config.Bucket, m.messageId)
	if err != nil {
		return fmt.Errorf("S3 error: %s", err.Error())
	}
	text, replyTo, err := getText(body)
	if err != nil {
		return err
	}

//	fmt.Fprintf(os.Stderr, "%s\n", text)

	if len(m.headers.from) > 1 {
		fmt.Fprintf(os.Stderr, "multiple From not supported: %v\n", m.headers.from)
	}
	from := m.headers.from[0] // only accept single sender
	if !isAuthSender(from) {
		return fmt.Errorf("sender is not in whitelist: %s", from)
	}
	if replyTo != "" {
		replyTo = m.headers.messageId
	}

	err = sendEmail(sess, from, text, m.headers.subject, replyTo)
	if err != nil {
		return fmt.Errorf("Error sending email: %s", err)
	}
	return nil
}

// eventToMail: unmarshal record from event
func eventToMail(event json.RawMessage) (mail, error) {
	var evs struct {
		Records []struct {
			EventSource string `json:"eventSource"`
			Ses         struct {
				Mail struct {
					MessageId     string `json:"messageID"`
					CommonHeaders struct {
						ReturnPath string   `json:"returnPath"`
						From       []string `json:"from"`
						Date       string   `json:"date"`
						To         []string `json:"to"`
						MessageId  string   `json:"messageId"`
						Subject    string   `json:"subject"`
					}
				}
			}
		}
	}
	var m mail
	if err := json.Unmarshal(event, &evs); err != nil {
		return mail{}, err
	}
	if len(evs.Records) > 1 {
		fmt.Fprintf(os.Stderr, "multiple records unexpected: %d\n", len(evs.Records))
	}
	m.messageId = evs.Records[0].Ses.Mail.MessageId
	m.headers.from = evs.Records[0].Ses.Mail.CommonHeaders.From
	m.headers.subject = evs.Records[0].Ses.Mail.CommonHeaders.Subject
	m.headers.messageId = evs.Records[0].Ses.Mail.CommonHeaders.MessageId
	return m, nil
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
func getText(body []byte) (string, string, error) {
	reader := bytes.NewReader(body)
	msg, err := email.ParseMessage(reader)
	if err != nil {
		return "", "", err
	}
	replyTo := msg.Header.Get("In-Reply-To")

	var text string
	msgBit := msg.PartsContentTypePrefix("text/plain")
	if len(msgBit) > 0 {
		for _, part := range msgBit {
			text = string(part.Body)
			// fmt.Println(part.Header["Content-Type"])
			// todo: parse "[text/plain; charset=UTF-8]"

		}
	} else {
		text = string(msg.Body) //
	}

	return text, replyTo, nil
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
	for _, valid := range config.Whitelist {
		if addr == valid {
			return true
		}
	}
	return false
}

// sendEmail: send email (and print response to stderr)
func sendEmail(sess *session.Session, from, text, subject, replyTo string) error {
	svc := ses.New(sess)
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ses/#example_SES_SendEmail
	params := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{ // Required
			Data: payload(from, text, subject, replyTo),
		},
		// Destinations: []*string{
		// 	aws.String("undisclosed recipient:"), // Required
		// 	// More values...
		// },
		Source: aws.String(config.InboxEmail),
	}
	resp, err := svc.SendRawEmail(params)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%v\n", resp)
	return nil
}

func payload(from, text, subject, messageId string) []byte { //
	var buf bytes.Buffer

	to := whitelistPtrs()
	for _, s := range to {
		if !strings.Contains(from, *s) {
                	buf.Write(header("Bcc", *s))
		}
	}
	//buf.WriteString(nl)
        //build from
        index := strings.Index(from, "<")
        if(index == -1) {
          index = strings.Index(from, "@")
        }
        fromText := from[0:index] + " <" + config.InboxEmail + ">"
        fmt.Fprintf(os.Stderr, "From: %s\n", fromText)

	buf.Write(header("From", fromText))
	buf.Write(header("Reply-To", config.InboxEmail))
	buf.Write(header("Subject", subject))
	buf.Write(header("MIMIE-Version", "1.0"))
	buf.Write(header("Content-Type", "text/plain; charset=UTF-8"))
	buf.Write(header("In-Reply-To", messageId))

	buf.WriteString(nl)
	buf.WriteString(text)
	buf.WriteString(nl)

	return buf.Bytes()
}

func header(front, back string) []byte {
	var buf bytes.Buffer
	buf.WriteString(front)
	buf.WriteString(": ")
	buf.WriteString(back)
	buf.WriteString(nl)
	return buf.Bytes()
}

// return slice of pointers to whitelisted email addresses
// used to initialize ses.SendEmailInput data structure
func whitelistPtrs() []*string {
	to := make([]*string, len(config.Whitelist))
	for i := range config.Whitelist {
		to[i] = &config.Whitelist[i]
	}
	return to
}
