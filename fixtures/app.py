import requests, os, logging, sys
from lib.aws_auth import AwsAuth
from lib.jwt import generate_jwt

from flask import Flask, render_template, request

app = Flask(__name__, static_url_path="/assets")

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

handler = logging.StreamHandler(sys.stdout)
handler.setLevel(logging.DEBUG)
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger.addHandler(handler)


@app.route("/", methods=["GET", "POST"])
def gt_test():
    aws_auth = AwsAuth()

    uid = request.form.get("uid", "")
    json_data = request.form.get("json-data", "{}")
    base_url = os.environ["BASE_URL"]

    template_data = {
        "base_url": base_url,
        "uid": uid,
        "json_data": json_data,
    }

    if request.method == "GET":
        return render_template("index.html", **template_data)

    if request.method == "POST":
        url = base_url + "lpas/" + uid

        if aws_auth.is_authed:
            headers = aws_auth.get_headers(method="PUT", url=url, data=json_data)
        else:
            headers = {}

        token = generate_jwt("secret")

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
            **template_data,
            success=resp.status_code < 400,
            error=resp.text,
        )

    return "error"


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=80)
