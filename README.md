# ಇನ್ನೂ ಎಷ್ಟು ದಿನ? (innu eshtu dina?)

**Bengaluru's public scoreboard of promised-versus-delivered infrastructure.**

Every stalled project gets a live counter: days since it began, days past its
promised deadline, deadlines missed. No opinions, no adjectives — dates,
sources, and arithmetic. Delivered projects move to the **Hall of Finally**,
because fair is fair.

## How it works

- `data/projects/*.yaml` — one file per project. The dataset **is** the repo.
- `main.go` — static-site generator. Reads YAML, writes `public/`:
  HTML pages, plus a 1200×630 share image per project with the day-count
  baked in (regenerated daily by CI, so shared links always show today's
  number).
- Counters on the site tick live in the browser (IST).

## Run locally

```bash
go mod tidy
go run . -serve   # builds public/ and serves on http://localhost:8791
```

## Add a project

Copy any file in `data/projects/`, fill it in, open a PR. The bar: **every
date must trace to a linked news report or government document.** No source,
no listing.

```yaml
id: my-project            # slug, becomes /p/my-project/
name: Project Name
agency: BBMP              # agency only — never individuals
status: stalled           # stalled | crawling | resumed | completed
started: "2017-05-01"
started_approx: true      # if only month/year is public
promised:                 # every publicly promised completion date, in order
  - "2019-05-01"
  - "2026-10-31"
completed: ""             # set when delivered — project moves to Hall of Finally
cost_sanctioned_cr: 204   # optional, ₹ crore, sourced
summary: One paragraph, neutral.
notes:
  - What happened, one fact per line, each traceable to a source below.
sources:
  - title: "Headline (Publication, year)"
    url: https://...
```

## Deploy

GitHub Actions builds daily at 02:00 IST and deploys `public/` to Cloudflare
Pages. One-time setup: create a Pages project named `innu-eshtu-dina`, add
`CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` repo secrets, and set
`BASE_URL` in `.github/workflows/build.yml` to your domain.

## Corrections

Open an issue with a source. Corrections answered within 48 hours.

## License

Code: MIT. Data: CC BY 4.0 — reuse freely with attribution.
