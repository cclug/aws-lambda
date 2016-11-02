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
	bucket     = "cclug"
	inboxEmail = "CCLUG mailing list <inbox@email.cclug.org.au>"
	nl         = "\r\n"
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
	sess := session.Must(session.NewSession())
	body, err := getBody(sess, bucket, m.messageId)
	if err != nil {
		return fmt.Errorf("S3 error: %s", err.Error())
	}
	text, replyTo, err := getText(body)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s\n", text)

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
		return err
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
	for _, valid := range whitelist {
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
		Source: aws.String(from),
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

	buf.WriteString("Bcc: ")
	to := whitelistPtrs()
	end := len(to) - 1
	for i, s := range to {
		if strings.Contains(from, *s) {
			end--
			continue
		}
		buf.WriteString(*s)
		if i != end {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(nl)

	buf.Write(header("From", from))
	buf.Write(header("Reply-To", inboxEmail))
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
	whitelist := []string{"mkuchin@gmail.com", "me@tobin.cc"} // remove this for production
	to := make([]*string, len(whitelist))
	for i := range whitelist {
		to[i] = &whitelist[i]
	}
	return to
}
