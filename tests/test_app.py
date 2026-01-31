import pytest
from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as client:
        yield client


def test_index_page(client):
    """Test that the index page loads correctly."""
    response = client.get("/")
    assert response.status_code == 200
    assert b"Brave Search" in response.data or b"search" in response.data.lower()


def test_search_redirect_empty_query(client):
    """Test that empty query redirects to index."""
    response = client.get("/search?q=")
    assert response.status_code == 302
    assert response.headers["Location"] == "/"


def test_search_success(client, mocker):
    """Test a successful search with mocked API response."""
    mock_response = {
        "web": {
            "results": [
                {
                    "title": "Test Result",
                    "url": "https://example.com",
                    "description": "Description",
                }
            ]
        }
    }
    mocker.patch("app.search_brave", return_value=mock_response)

    response = client.get("/search?q=test")
    assert response.status_code == 200
    assert b"Test Result" in response.data
    assert b"https://example.com" in response.data
