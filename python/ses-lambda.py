import json
import boto3
import urlparse
import urllib
import urllib2

# required config
EMAIL = 'xxx@email.com'
 
# optional config
CAPTCHA_ENABLED = False
DEBUG = False
CAPTCHA_SECRET = None

# const
CAPTCHA_FIELD = 'g-recaptcha-response'
CAPTCHA_API = 'https://www.google.com/recaptcha/api/siteverify'
SUBJECT = 'contact form'

def checkCaptcha(response):
  values = {'secret' : CAPTCHA_SECRET,
            'response' : response}
  data = urllib.urlencode(values)
  req = urllib2.Request(CAPTCHA_API, data)
  response = urllib2.urlopen(req)
  response = response.read()
  return json.loads(response)['success']

def sendEmail(message):
  client = boto3.client('ses', region_name='us-west-2')
  response = client.send_email(
    Source=EMAIL,
    Destination = {
      'ToAddresses': [EMAIL]
    }, Message = {
      'Subject': {'Data': SUBJECT},
      'Body': { 'Text': { 'Data': message } }
    })
  return response

def lambda_handler(event, context):
    params = urlparse.parse_qs(event['body'])
    human = True
    if CAPTCHA_ENABLED:
       if CAPTCHA_FIELD in params:
         human = checkCaptcha(params[CAPTCHA_FIELD][0])
       else: 
         human = False

    body = "From: %s\nMessage:\n%s" % (params['email'][0], params['message'][0])

    if DEBUG:
      eventJson = json.dumps(event, indent=4, sort_keys=True)
      body += "\n-------------------\nEvent json: %s\n" % eventJson
    
    if human:
      response = sendEmail(body)
      if 'redirect' in params:
        return {"statusCode": 302, "headers": {"Location": params['redirect'][0]}, "body": ''}
      else:
        responseJson = json.dumps(response, indent=4, sort_keys=True)  
        return {"statusCode": 200, "headers": {}, "body": responseJson}
    else:
      return {"statusCode": 200, "headers": {}, "body": "Bot detected!"}
