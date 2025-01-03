import requests, os, logging, sys, json, uuid
from lib.aws_auth import AwsAuth
from lib.jwt import generate_jwt
from urllib.parse import quote

from flask import Flask, render_template, request, jsonify
from flask_wtf import CSRFProtect

app = Flask(__name__, static_url_path="/assets")
app.config.update(
    SECRET_KEY=uuid.uuid4().__str__(),
)

csrf = CSRFProtect()
csrf.init_app(app)

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

handler = logging.StreamHandler(sys.stdout)
handler.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger.addHandler(handler)


@app.route("/health-check", methods=["GET"])
def health_check():
    return jsonify({"ok": True})


@app.route("/health-check/service", methods=["GET"])
@app.route("/health-check/dependencies", methods=["GET"])
def health_check_dependencies():
    try:
        aws_auth = AwsAuth()
        url = os.environ["BASE_URL"] + "/health-check"

        if aws_auth.is_authed:
            headers = aws_auth.get_headers(method="GET", url=url)
        else:
            headers = {}

        requests.get(
            url,
            headers={
                **headers,
                "Content-Type": "application/json",
            },
        )

        return jsonify({"ok": True})

    except Exception as e:
        logger.error("healthcheck failed: " + e.__class__.__name__, {"exception": e})

        return jsonify({"ok": False})


@app.route("/", methods=["GET"])
def get_index():
    base_url = os.environ["BASE_URL"]

    return render_template(
        "index.html",
        **{
            "base_url": base_url,
            "json_data": "{}",
        },
    )


@app.route("/", methods=["POST"])
def post_index():
    aws_auth = AwsAuth()

    uid = request.form.get("uid", "")
    json_data = request.form.get("json-data", "{}")
    base_url = os.environ["BASE_URL"]

    url = base_url + "/lpas/" + quote(uid)

    if aws_auth.is_authed:
        headers = aws_auth.get_headers(method="PUT", url=url, data=json_data)
    else:
        headers = {}

    token = generate_jwt(os.environ["JWT_SECRET_KEY"])

    resp = requests.put(
        url,
        json_data,
        headers={
            **headers,
            "Content-Type": "application/json",
            "X-Jwt-Authorization": "Bearer " + token,
        },
    )

    return render_template(
        "index.html",
        **{
            "base_url": base_url,
            "uid": uid,
            "json_data": json_data,
        },
        success=resp.status_code < 400,
        error=json.loads(resp.text),
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=80)
