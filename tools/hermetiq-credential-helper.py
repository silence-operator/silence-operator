#!/usr/bin/env python3
import json
import os
import sys
import urllib.request



CLIENT_ID = os.getenv("CLIENT_ID")
CLIENT_SECRET = os.getenv("CLIENT_SECRET")
TOKEN_URL = "https://login.cloud.hermetiq.io/v1/oauth2/token"

try:
    data = json.dumps({"client_id": CLIENT_ID, "client_secret": CLIENT_SECRET, "grant_type": "client_credentials"}).encode()
    request = urllib.request.Request(TOKEN_URL, data=data, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(request, timeout=30) as response:
        token = json.load(response)["access_token"]
    if not isinstance(token, str) or not token.strip():
        raise ValueError("The token response did not include a usable access token")
    print(json.dumps({"headers": {"x-hermetiq-m2m-access-token": [token]}}))
except Exception:
    print("Unable to obtain credentials. Contact your administrator.", file=sys.stderr)
    sys.exit(1)
