import json, requests, os
from lib.aws_auth import AwsAuth
from lib.jwt import generate_jwt
from urllib.parse import quote


def __get_lpa(uid):
    aws_auth = AwsAuth()

    base_url = os.environ["BASE_URL"]

    url = base_url + "/lpas/" + quote(uid)

    if aws_auth.is_authed:
        headers = aws_auth.get_headers(method="GET", url=url)
    else:
        headers = {}

    token = generate_jwt(os.environ["JWT_SECRET_KEY"])

    response = requests.get(
        url,
        headers={
            **headers,
            "Content-Type": "application/json",
            "X-Jwt-Authorization": "Bearer " + token,
        },
    )

    return json.loads(response.text)


def __send_update(uid, type, *changes) -> requests.Response:
    aws_auth = AwsAuth()

    base_url = os.environ["BASE_URL"]

    url = base_url + "/lpas/" + quote(uid) + "/updates"

    body = json.dumps(
        {
            "type": type,
            "changes": changes,
        }
    )

    if aws_auth.is_authed:
        headers = aws_auth.get_headers(method="POST", url=url, data=body)
    else:
        headers = {}

    token = generate_jwt(os.environ["JWT_SECRET_KEY"])

    return requests.post(
        url,
        body,
        headers={
            **headers,
            "Content-Type": "application/json",
            "X-Jwt-Authorization": "Bearer " + token,
        },
    )


def attorney_sign(uid, attorney_uuid, signed_at) -> requests.Response:
    lpa = __get_lpa(uid)

    attorney_index = None
    for index, attorney in enumerate(lpa["attorneys"]):
        if attorney["uid"] == attorney_uuid:
            attorney_index = index

    if attorney_index == None:
        raise Exception("Could not find attorney with UID {}".format(attorney_uuid))

    return __send_update(
        uid,
        "ATTORNEY_SIGN",
        {
            "key": "/attorneys/{}/signedAt".format(attorney_index),
            "old": None,
            "new": signed_at,
        },
    )


def certificate_provider_sign(uid, signed_at) -> requests.Response:
    return __send_update(
        uid,
        "CERTIFICATE_PROVIDER_SIGN",
        {
            "key": "/certificateProvider/signedAt",
            "old": None,
            "new": signed_at,
        },
    )

def donor_id(uid, checked_at, id_type) -> requests.Response:
    return __send_update(
        uid,
        "DONOR_CONFIRM_IDENTITY",
        {
            "key": "/donor/identityCheck/checkedAt",
            "old": None,
            "new": checked_at,
        },
        {
            "key": "/donor/identityCheck/type",
            "old": None,
            "new": id_type,
        },
    )

def certificate_provider_id(uid, checked_at, type) -> requests.Response:
    return __send_update(
        uid,
        "CERTIFICATE_PROVIDER_CONFIRM_IDENTITY",
        {
            "key": "/certificateProvider/identityCheck/checkedAt",
            "old": None,
            "new": checked_at,
        },
        {
            "key": "/certificateProvider/identityCheck/type",
            "old": None,
            "new": type,
        },
    )
