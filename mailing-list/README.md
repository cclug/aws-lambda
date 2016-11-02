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

## Setting up DNS to host lambda.cclug.org.au zone on Route 53
```
lambda.cclug.org.au. 21599	IN	NS	ns-1333.awsdns-38.org.
lambda.cclug.org.au. 21599	IN	NS	ns-148.awsdns-18.com.
lambda.cclug.org.au. 21599	IN	NS	ns-1976.awsdns-55.co.uk.
lambda.cclug.org.au. 21599	IN	NS	ns-572.awsdns-07.net.
```
