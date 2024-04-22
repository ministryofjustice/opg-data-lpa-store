import jwt, math, time


def generate_jwt(secret, sub="someone@someplace.somewhere.com"):
    return jwt.encode(
        {
            "exp": math.floor(time.time() + 60 * 65),
            "iat": math.floor(time.time() - 60 * 65),
            "iss": "opg.poas.sirius",
            "sub": sub,
        },
        secret,
        algorithm="HS256",
    )
