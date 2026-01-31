from flask import Flask, render_template, request, redirect, url_for
from brave_search import search_brave

app = Flask(__name__)


@app.route("/")
def index():
    return render_template("index.html")


@app.route("/search")
def search():
    query = request.args.get("q")
    if not query:
        return redirect(url_for("index"))

    try:
        results = search_brave(query)
        web_results = results.get("web", {}).get("results", [])
        return render_template("results.html", query=query, results=web_results)
    except Exception as e:
        return f"Error: {str(e)}", 500


if __name__ == "__main__":
    app.run(debug=True, port=5001)
