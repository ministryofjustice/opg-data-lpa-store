import jwt, math, time


def generate_jwt(secret, sub="urn:opg:sirius:users:100"):
    return jwt.encode(
        {
            "exp": math.floor(time.time() + 60 * 5),
            "iat": math.floor(time.time()),
            "iss": "opg.poas.sirius",
            "sub": sub,
        },
        secret,
        algorithm="HS256",
    )
