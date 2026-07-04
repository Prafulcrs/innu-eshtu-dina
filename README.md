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

## Deploy (Vercel)

1. Push this repo to GitHub.
2. Import it at [vercel.com/new](https://vercel.com/new) — framework
   "Other". `vercel.json` already sets the build command (`go run .`) and
   output directory (`public`). Every push to `main` deploys.
3. Keep share images fresh: in the Vercel project, create a Deploy Hook
   (Settings → Git → Deploy Hooks), add its URL as the `VERCEL_DEPLOY_HOOK`
   repo secret on GitHub. The included workflow triggers a rebuild daily at
   02:00 IST so OG day-counts stay current.
4. Custom domain: add it in Vercel, then set `BASE_URL` (Vercel project
   env var) to the domain so OG image URLs match.

## Corrections

Open an issue with a source. Corrections answered within 48 hours.

## License

Code: MIT. Data: CC BY 4.0 — reuse freely with attribution.
