# Mailing List in Golang running on AWS Lambda

**Requires [Apex](apex.run)**

See functions/email/main.go for main function.
See functions/email/main_test.go for example local testing.

## Usage

To run the example first setup your
[AWS Credentials](http://apex.run/#aws-credentials), and ensure "role" in
./project.json is set to your role ARN.

Deploy the function:

`$ apex deploy email`

## Setting up DNS to host email.cclug.org.au zone on Route 53
```
email.cclug.org.au.	172800	IN	NS	ns-1531.awsdns-63.org.
email.cclug.org.au.	172800	IN	NS	ns-1897.awsdns-45.co.uk.
email.cclug.org.au.	172800	IN	NS	ns-407.awsdns-50.com.
email.cclug.org.au.	172800	IN	NS	ns-544.awsdns-04.net.
```
