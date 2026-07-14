// innu eshtu dina? — Bengaluru's infrastructure scoreboard.
// Static-site generator: reads data/projects/*.yaml, writes public/.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.tmpl
var tmplFS embed.FS

var ist = time.FixedZone("IST", 5*3600+30*60)

type Source struct {
	Title string `yaml:"title"`
	URL   string `yaml:"url"`
}

type Project struct {
	ID            string   `yaml:"id"`
	Name          string   `yaml:"name"`
	Agency        string   `yaml:"agency"`
	Status        string   `yaml:"status"`
	Started       string   `yaml:"started"`
	StartedApprox bool     `yaml:"started_approx"`
	Promised      []string `yaml:"promised"`
	Completed     string   `yaml:"completed"`
	CostCr        float64  `yaml:"cost_sanctioned_cr"`
	Summary       string   `yaml:"summary"`
	Notes         []string `yaml:"notes"`
	Sources       []Source `yaml:"sources"`
	Location      struct {
		Lat float64 `yaml:"lat"`
		Lng float64 `yaml:"lng"`
	} `yaml:"location"`

	// derived
	StartedT     time.Time
	PromisedT    []time.Time
	CompletedT   time.Time
	StatusLabel  string
	CounterFrom  string // date the client-side counter ticks from
	CounterDays  int
	CounterLabel string
	PromiseLine  string
	LateDays     int // completed projects: days past first promise
	Timeline     []TimelineRow
	ChartDots    []ChartDot
	ChartToday   float64
}

type ChartDot struct {
	X     float64
	Class string
	Label string
}

type TimelineRow struct {
	Date   string
	Label  string
	Detail string
	Class  string
}

type Page struct {
	Title   string
	Desc    string
	OGImage string
	Year    int

	// index
	Active      []*Project
	Done        []*Project
	TotProjects int
	TotOverdue  int
	TotCost     float64
	MapJS       template.JS

	// project page
	P *Project
}

var statusLabels = map[string]string{
	"stalled":   "Stalled",
	"crawling":  "Crawling",
	"resumed":   "Work resumed",
	"completed": "Done. Finally.",
}

func main() {
	serve := flag.Bool("serve", false, "serve public/ on :8791 after building")
	flag.Parse()

	now := time.Now().In(ist)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://innu-eshtu-dina.vercel.app"
	}

	projects, err := loadProjects("data/projects", now)
	if err != nil {
		log.Fatalf("load projects: %v", err)
	}
	if err := build(projects, baseURL, now); err != nil {
		log.Fatalf("build: %v", err)
	}
	log.Printf("built %d project pages into public/", len(projects))

	if *serve {
		log.Println("serving on http://localhost:8791")
		log.Fatal(http.ListenAndServe(":8791", http.FileServer(http.Dir("public"))))
	}
}

func loadProjects(dir string, now time.Time) ([]*Project, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var projects []*Project
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var p Project
		if err := yaml.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if err := derive(&p, now); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		projects = append(projects, &p)
	}
	return projects, nil
}

func mustDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, ist)
}

func daysBetween(from, to time.Time) int {
	d := int(to.Sub(from).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

func derive(p *Project, now time.Time) error {
	var err error
	if p.StartedT, err = mustDate(p.Started); err != nil {
		return fmt.Errorf("started: %w", err)
	}
	for _, s := range p.Promised {
		t, err := mustDate(s)
		if err != nil {
			return fmt.Errorf("promised: %w", err)
		}
		p.PromisedT = append(p.PromisedT, t)
	}
	if p.Completed != "" {
		if p.CompletedT, err = mustDate(p.Completed); err != nil {
			return fmt.Errorf("completed: %w", err)
		}
	}
	label, ok := statusLabels[p.Status]
	if !ok {
		return fmt.Errorf("unknown status %q", p.Status)
	}
	p.StatusLabel = label

	fmonth := func(t time.Time) string { return t.Format("Jan 2006") }

	// counters + promise line
	switch {
	case p.Status == "completed":
		first := p.PromisedT[0]
		p.LateDays = daysBetween(first, p.CompletedT)
		p.CounterFrom = p.Promised[0]
		p.CounterDays = p.LateDays
		p.CounterLabel = "days late, but done"
		p.PromiseLine = fmt.Sprintf("Promised %s · opened %s", fmonth(first), fmonth(p.CompletedT))

	case len(p.PromisedT) > 0 && p.PromisedT[0].Before(now):
		p.CounterFrom = p.Promised[0]
		p.CounterDays = daysBetween(p.PromisedT[0], now)
		p.CounterLabel = "days past the first promised deadline"
		missed := 0
		for _, t := range p.PromisedT {
			if t.Before(now) {
				missed++
			}
		}
		plural := "deadlines"
		if missed == 1 {
			plural = "deadline"
		}
		last := p.PromisedT[len(p.PromisedT)-1]
		if last.After(now) {
			p.PromiseLine = fmt.Sprintf("%d %s missed · current target %s", missed, plural, fmonth(last))
		} else {
			p.PromiseLine = fmt.Sprintf("%d %s missed · no fresh official date", missed, plural)
		}

	default:
		p.CounterFrom = p.Started
		p.CounterDays = daysBetween(p.StartedT, now)
		p.CounterLabel = "days since the project began"
		if len(p.PromisedT) > 0 {
			p.PromiseLine = fmt.Sprintf("Announced %s · current target %s",
				fmonth(p.StartedT), fmonth(p.PromisedT[len(p.PromisedT)-1]))
		} else {
			p.PromiseLine = fmt.Sprintf("Announced %s · no completion date ever promised", fmonth(p.StartedT))
		}
	}

	// timeline
	startDetail := ""
	if p.StartedApprox {
		startDetail = "Date approximate — see methodology"
	}
	p.Timeline = append(p.Timeline, TimelineRow{
		Date: p.StartedT.Format("2 Jan 2006"), Label: "Project began", Detail: startDetail,
	})
	for i, t := range p.PromisedT {
		isLast := i == len(p.PromisedT)-1
		switch {
		case p.Status == "completed":
			p.Timeline = append(p.Timeline, TimelineRow{
				Date: t.Format("2 Jan 2006"), Label: "Promised completion", Detail: "Missed", Class: "missed",
			})
		case t.Before(now):
			p.Timeline = append(p.Timeline, TimelineRow{
				Date: t.Format("2 Jan 2006"), Label: "Promised completion",
				Detail: fmt.Sprintf("Missed — %s days ago", comma(daysBetween(t, now))), Class: "missed",
			})
		case isLast:
			p.Timeline = append(p.Timeline, TimelineRow{
				Date: t.Format("2 Jan 2006"), Label: "Current official target",
				Detail: fmt.Sprintf("%s days away", comma(daysBetween(now, t))), Class: "target",
			})
		default:
			p.Timeline = append(p.Timeline, TimelineRow{
				Date: t.Format("2 Jan 2006"), Label: "Promised completion", Class: "target",
			})
		}
	}
	if p.Status == "completed" {
		p.Timeline = append(p.Timeline, TimelineRow{
			Date: p.CompletedT.Format("2 Jan 2006"), Label: "Opened",
			Detail: fmt.Sprintf("%s days after the first promised deadline", comma(p.LateDays)), Class: "done",
		})
	}

	// promise-drift strip: dots on a time axis from start to the latest known date
	end := now
	if len(p.PromisedT) > 0 {
		if last := p.PromisedT[len(p.PromisedT)-1]; last.After(end) {
			end = last
		}
	}
	if !p.CompletedT.IsZero() && p.CompletedT.After(end) {
		end = p.CompletedT
	}
	span := end.Sub(p.StartedT).Seconds()
	xpos := func(t time.Time) float64 {
		return 20 + 960*t.Sub(p.StartedT).Seconds()/span
	}
	p.ChartToday = xpos(now)
	p.ChartDots = append(p.ChartDots, ChartDot{X: 20, Class: "dot-start", Label: "Began " + p.StartedT.Format("Jan 2006")})
	for _, t := range p.PromisedT {
		class := "dot-target"
		if t.Before(now) || p.Status == "completed" {
			class = "dot-missed"
		}
		p.ChartDots = append(p.ChartDots, ChartDot{X: xpos(t), Class: class, Label: "Promised " + t.Format("Jan 2006")})
	}
	if p.Status == "completed" {
		p.ChartDots = append(p.ChartDots, ChartDot{X: xpos(p.CompletedT), Class: "dot-done", Label: "Opened " + p.CompletedT.Format("Jan 2006")})
	}
	return nil
}

// comma formats an integer with Indian digit grouping (12,34,567).
func comma(n int) string {
	s := fmt.Sprintf("%d", n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	if len(s) > 3 {
		head, tail := s[:len(s)-3], s[len(s)-3:]
		var parts []string
		for len(head) > 2 {
			parts = append([]string{head[len(head)-2:]}, parts...)
			head = head[:len(head)-2]
		}
		if head != "" {
			parts = append([]string{head}, parts...)
		}
		s = strings.Join(parts, ",") + "," + tail
	}
	if neg {
		s = "-" + s
	}
	return s
}

func build(projects []*Project, baseURL string, now time.Time) error {
	if err := os.RemoveAll("public"); err != nil {
		return err
	}
	for _, d := range []string{"public/static", "public/og", "public/methodology"} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	// static assets
	staticFiles, err := os.ReadDir("static")
	if err != nil {
		return err
	}
	for _, f := range staticFiles {
		if f.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("static", f.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join("public/static", f.Name()), raw, 0o644); err != nil {
			return err
		}
	}

	funcs := template.FuncMap{
		"comma":  comma,
		"commaf": func(f float64) string { return comma(int(f)) },
		"fdate":  func(t time.Time) string { return t.Format("2 Jan 2006") },
		"inc":    func(i int) int { return i + 1 },
	}
	tmpl, err := template.New("site").Funcs(funcs).ParseFS(tmplFS, "templates/*.tmpl")
	if err != nil {
		return err
	}

	var active, done []*Project
	totOverdue, totCost := 0, 0.0
	for _, p := range projects {
		totCost += p.CostCr
		if p.Status == "completed" {
			done = append(done, p)
			totOverdue += p.LateDays
			continue
		}
		active = append(active, p)
		if p.CounterLabel == "days past the first promised deadline" {
			totOverdue += p.CounterDays
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].CounterDays > active[j].CounterDays })
	sort.Slice(done, func(i, j int) bool { return done[i].LateDays > done[j].LateDays })

	render := func(outPath, tmplName string, page Page) error {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		return tmpl.ExecuteTemplate(f, tmplName, page)
	}

	year := now.Year()

	// map markers for the homepage
	type marker struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Lat   float64 `json:"lat"`
		Lng   float64 `json:"lng"`
		Days  int     `json:"days"`
		Label string  `json:"label"`
	}
	var markers []marker
	for _, p := range projects {
		if p.Location.Lat == 0 {
			continue
		}
		days, label := p.CounterDays, p.CounterLabel
		if p.Status == "completed" {
			days, label = p.LateDays, "days late, but done"
		}
		markers = append(markers, marker{p.ID, p.Name, p.Location.Lat, p.Location.Lng, days, label})
	}
	mapJSON, err := json.Marshal(markers)
	if err != nil {
		return err
	}

	// index
	if err := render("public/index.html", "index", Page{
		MapJS: template.JS(mapJSON),
		Title:   "innu eshtu dina? — Bengaluru's infrastructure scoreboard",
		Desc:    fmt.Sprintf("%d Bengaluru projects, %s combined days past promised deadlines. Live counters, sourced dates, plain arithmetic.", len(projects), comma(totOverdue)),
		OGImage: baseURL + "/og/site.png",
		Year:    year, Active: active, Done: done,
		TotProjects: len(projects), TotOverdue: totOverdue, TotCost: totCost,
	}); err != nil {
		return err
	}

	// project pages + OG images
	for _, p := range projects {
		if err := render(filepath.Join("public/p", p.ID, "index.html"), "project", Page{
			Title:   p.Name + " — innu eshtu dina?",
			Desc:    fmt.Sprintf("%s. %s: %s.", p.PromiseLine, p.Agency, p.StatusLabel),
			OGImage: baseURL + "/og/" + p.ID + ".png",
			Year:    year, P: p,
		}); err != nil {
			return err
		}
		if err := drawProjectOG(p, filepath.Join("public/og", p.ID+".png")); err != nil {
			return err
		}
	}

	// methodology
	if err := render("public/methodology/index.html", "methodology", Page{
		Title:   "Methodology — innu eshtu dina?",
		Desc:    "How the counters are computed, and the sourcing rules behind every number.",
		OGImage: baseURL + "/og/site.png",
		Year:    year,
	}); err != nil {
		return err
	}

	return drawSiteOG(totOverdue, len(projects), "public/og/site.png")
}

// --- OG images ---

func face(ttf []byte, size float64) font.Face {
	f, err := truetype.Parse(ttf)
	if err != nil {
		log.Fatalf("parse font: %v", err)
	}
	return truetype.NewFace(f, &truetype.Options{Size: size})
}

const (
	ogW, ogH  = 1200, 630
	ogBg      = "#0d1117"
	ogAccent  = "#ffd23f"
	ogRed     = "#f85149"
	ogGreen   = "#3fb950"
	ogMuted   = "#8b949e"
	ogText    = "#e6edf3"
	ogBrand   = "INNU ESHTU DINA?  ·  Bengaluru's infrastructure scoreboard"
)

func ogCanvas() *gg.Context {
	dc := gg.NewContext(ogW, ogH)
	dc.SetHexColor(ogBg)
	dc.Clear()
	dc.SetHexColor(ogAccent)
	dc.DrawRectangle(0, 0, ogW, 10)
	dc.Fill()
	return dc
}

func drawProjectOG(p *Project, outPath string) error {
	dc := ogCanvas()

	num, label, color := p.CounterDays, strings.ToUpper(p.CounterLabel), ogRed
	if p.Status == "completed" {
		num, label, color = p.LateDays, "DAYS LATE — BUT DONE", ogGreen
	}

	dc.SetFontFace(face(gobold.TTF, 200))
	dc.SetHexColor(color)
	dc.DrawStringAnchored(comma(num), ogW/2, 210, 0.5, 0.5)

	dc.SetFontFace(face(goregular.TTF, 30))
	dc.SetHexColor(ogMuted)
	dc.DrawStringAnchored(label, ogW/2, 350, 0.5, 0.5)

	dc.SetFontFace(face(gobold.TTF, 44))
	dc.SetHexColor(ogText)
	dc.DrawStringWrapped(p.Name, ogW/2, 420, 0.5, 0, 1080, 1.25, gg.AlignCenter)

	dc.SetFontFace(face(goregular.TTF, 26))
	dc.SetHexColor(ogAccent)
	dc.DrawStringAnchored(ogBrand, ogW/2, 585, 0.5, 0.5)

	return dc.SavePNG(outPath)
}

func drawSiteOG(totOverdue, nProjects int, outPath string) error {
	dc := ogCanvas()

	dc.SetFontFace(face(gobold.TTF, 200))
	dc.SetHexColor(ogAccent)
	dc.DrawStringAnchored(comma(totOverdue), ogW/2, 210, 0.5, 0.5)

	dc.SetFontFace(face(goregular.TTF, 30))
	dc.SetHexColor(ogMuted)
	dc.DrawStringAnchored("COMBINED DAYS PAST PROMISED DEADLINES", ogW/2, 350, 0.5, 0.5)

	dc.SetFontFace(face(gobold.TTF, 44))
	dc.SetHexColor(ogText)
	dc.DrawStringAnchored(fmt.Sprintf("%d Bengaluru projects, one scoreboard", nProjects), ogW/2, 440, 0.5, 0.5)

	dc.SetFontFace(face(goregular.TTF, 26))
	dc.SetHexColor(ogAccent)
	dc.DrawStringAnchored(ogBrand, ogW/2, 585, 0.5, 0.5)

	return dc.SavePNG(outPath)
}
