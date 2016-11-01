# Mailing List in Golang running on AWS Lambda

**Requires [Apex](apex.run)**

See functions/email/main.go for main function.
See functions/email/main_test.go for example local testing.

## Usage

To run the example first setup your
[AWS Credentials](http://apex.run/#aws-credentials), and ensure "role" in
./project.json is set to your role ARN.

Deploy the functions:

`$ apex deploy`

Try it out:

`$ apex invoke simple < event.json`
