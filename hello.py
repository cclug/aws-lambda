import json

def lambda_handler(event, context):
    
    eventJson = json.dumps(event, indent=4, sort_keys=True)
    body = "event: %s\n" % (eventJson)
    return {"statusCode": 200,
    "headers": {"Lambda-X": "test"},
    "body": body }
