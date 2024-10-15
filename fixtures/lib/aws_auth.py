import boto3
import os
from botocore import crt, awsrequest


class AwsAuth:
    def __init__(self, boto3_session=boto3.Session()):

        if os.environ.get("SKIP_AUTH", "0") == "1":
            self.is_authed = False
        else:
            self.is_authed = True
            self.session = boto3_session

    def get_headers(self, service = "execute-api", **request_config):
        sig_v4a = crt.auth.CrtS3SigV4AsymAuth(
            self.session.get_credentials(),
            service,
            os.environ.get("AWS_REGION", "eu-west-1"),
        )
        aws_req = awsrequest.AWSRequest(**request_config)
        sig_v4a.add_auth(aws_req)
        prepped = aws_req.prepare()

        return prepped.headers
