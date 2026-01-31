import os
import requests
from dotenv import load_dotenv

load_dotenv()

BRAVE_API_KEY = os.getenv("BRAVE_API_KEY")
BRAVE_SEARCH_URL = "https://api.search.brave.com/res/v1/web/search"


def search_brave(query):
    if not BRAVE_API_KEY:
        raise ValueError("BRAVE_API_KEY not found in environment variables")

    headers = {
        "Accept": "application/json",
        "Accept-Encoding": "gzip",
        "X-Subscription-Token": BRAVE_API_KEY,
    }
    params = {"q": query}

    response = requests.get(BRAVE_SEARCH_URL, headers=headers, params=params)
    response.raise_for_status()
    return response.json()
