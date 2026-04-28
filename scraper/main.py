from fastapi import FastAPI, Body
from fastapi.responses import PlainTextResponse
from bs4 import BeautifulSoup, Comment
import html2text

app = FastAPI()


def clean_html(html):
    soup = BeautifulSoup(html, "html.parser")
    for tag in soup(
        [
            "script", "style", "header", "footer", "noscript", "form",
            "input", "textarea", "select", "option", "button", "svg",
            "iframe", "object", "embed", "applet", "nav", "navbar",
        ]
    ):
        tag.decompose()

    for id_ in ["layers"]:
        tag = soup.find(id=id_)
        if tag:
            tag.decompose()

    for tag in soup.find_all(True):
        tag.attrs = {
            k: v for k, v in tag.attrs.items() if k not in ["class", "id", "style"]
        }

    for comment in soup.find_all(string=lambda text: isinstance(text, Comment)):
        comment.extract()

    return str(soup)


def parse_html_to_markdown(html):
    cleaned = clean_html(html)
    h = html2text.HTML2Text()
    h.ignore_links = False
    h.ignore_tables = False
    h.bypass_tables = False
    h.ignore_images = False
    h.protect_links = True
    h.mark_code = True
    content = h.handle(cleaned)
    soup = BeautifulSoup(html, "html.parser")
    title = soup.title.string if soup.title else "No title"
    return title, content


@app.post("/convert")
def convert(html_content: str = Body(default="", media_type="text/html")):
    if not html_content:
        return PlainTextResponse("Failed to retrieve content", status_code=400)
    title, content = parse_html_to_markdown(html_content)
    return PlainTextResponse(f"Title: {title}\n\nContent:\n{content}")
