# Weekly news digest: Google News RSS headlines for each tracked project.
# Output: digest.md (only sections with hits). Run by news-digest workflow.
import glob
import re
import urllib.parse
import urllib.request
import xml.etree.ElementTree as ET

names = []
for f in sorted(glob.glob("data/projects/*.yaml")):
    m = re.search(r"^name:\s*(.+)$", open(f, encoding="utf-8").read(), re.M)
    if m:
        names.append(m.group(1).strip().strip("'\""))

out = [
    "Weekly headlines mentioning tracked projects (last 7 days).",
    "If any deadline, status or amount changed, update the project YAML.",
    "",
]
for n in names:
    # ponytail: full names are too specific for news queries — search on the
    # first few words before any parenthesis, plus "Bengaluru"
    base = n.split("(")[0]
    words = re.sub(r"[^\w\s-]", " ", base).split()[:4]
    q = urllib.parse.quote('"%s" Bengaluru when:7d' % " ".join(words))
    url = f"https://news.google.com/rss/search?q={q}&hl=en-IN&gl=IN&ceid=IN:en"
    try:
        xml = urllib.request.urlopen(url, timeout=30).read()
        items = list(ET.fromstring(xml).iter("item"))[:5]
    except Exception:
        items = []
    if items:
        out.append(f"### {n}")
        for i in items:
            out.append(f"- [{i.findtext('title')}]({i.findtext('link')})")
        out.append("")

open("digest.md", "w", encoding="utf-8").write("\n".join(out))
