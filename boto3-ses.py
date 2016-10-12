import boto3

client = boto3.client('ses', region_name='us-west-2')
email = 'xxx@gmail.com'
subj = 'subj'
body = 'test'
response = client.send_email(
    Source=email,
    Destination={
        'ToAddresses': [email]
    },
    Message={
        'Subject': {'Data': subj},
        'Body': { 'Text': { 'Data': body } }
    })

print response
