import json
import boto3
import urlparse

def lambda_handler(event, context):
    params = urlparse.parse_qs(event['body'])
    
    eventJson = json.dumps(event, indent=4, sort_keys=True)
    body = "From: %s\nMessage:\n%s\nevent json: %s\n" % (params['email'][0], params['message'][0], eventJson)
   
    client = boto3.client('ses', region_name='us-west-2')
    email = 'xxxx@gmail.com'
    subj = 'api call'
    response = client.send_email(
      Source=email,
      Destination={
        'ToAddresses': [email]
      },
      Message={
        'Subject': {'Data': subj},
        'Body': { 'Text': { 'Data': body } }
      })
    
    responseJson = json.dumps(response, indent=4, sort_keys=True)
    return {"statusCode": 200, "headers": {}, "body": responseJson}
