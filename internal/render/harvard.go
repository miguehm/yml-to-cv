package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/miguehm/cvyml/assets/fonts"
	"github.com/miguehm/cvyml/internal/model"
)

const (
	fontFamily       = "Tinos"
	defaultPaper     = "letter"
	defaultMarginMM  = 19.05 // 0.75 inch
	defaultBodySize  = 10.0
	defaultNameSize  = 20.0
	defaultSectSize  = 11.5
	defaultDateSize  = 9.5
	fixedCreationUTC = 2000
)

// harvardRenderer implements Renderer.
type harvardRenderer struct {
	opts Options
}

func (o Options) withDefaults() Options {
	if o.Paper == "" {
		o.Paper = defaultPaper
	}
	if o.Margin <= 0 {
		o.Margin = defaultMarginMM
	}
	if o.BodySize <= 0 {
		o.BodySize = defaultBodySize
	}
	if o.CreationDate.IsZero() {
		o.CreationDate = time.Date(fixedCreationUTC, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return o
}

func (h *harvardRenderer) Render(w io.Writer, cv *model.CV) error {
	o := h.opts.withDefaults()

	pdf := fpdf.New("P", "mm", strings.ToLower(o.Paper), "")
	pdf.SetMargins(o.Margin, o.Margin, o.Margin)
	pdf.SetAutoPageBreak(true, o.Margin)
	pdf.SetCompression(true)
	pdf.SetCreationDate(o.CreationDate)
	pdf.SetModificationDate(o.CreationDate)
	pdf.SetTitle(cv.Basics.Name+" - CV", true)
	pdf.SetAuthor(cv.Basics.Name, true)

	pdf.AddUTF8FontFromBytes(fontFamily, "", fonts.TinosRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", fonts.TinosBold)
	pdf.AddUTF8FontFromBytes(fontFamily, "I", fonts.TinosItalic)
	pdf.AddUTF8FontFromBytes(fontFamily, "BI", fonts.TinosBoldItalic)

	pdf.AddPage()

	d := &harvardDoc{
		pdf:       pdf,
		o:         o,
		nameSize:  defaultNameSize,
		sectSize:  defaultSectSize,
		dateSize:  defaultDateSize,
		bodySize:  o.BodySize,
		lineH:     ptToMM(o.BodySize) * 1.28,
		indent:    4.0,
		left:      o.Margin,
		pageWidth: 0,
	}
	d.pageWidth, d.pageHeight = pdf.GetPageSize()
	d.contentWidth = d.pageWidth - 2*o.Margin
	d.right = d.pageWidth - o.Margin
	d.pageBreakTrigger = d.pageHeight - o.Margin

	d.header(cv.Basics)

	for _, name := range cv.SectionOrder {
		if !cv.HasSection(name) {
			continue
		}
		switch name {
		case model.SectionEducation:
			d.title("Education")
			for _, e := range cv.Education {
				d.education(e)
			}
		case model.SectionExperience:
			d.title("Experience")
			for _, e := range cv.Experience {
				d.experience(e)
			}
		case model.SectionProjects:
			d.title("Projects")
			for _, p := range cv.Projects {
				d.project(p)
			}
		case model.SectionSkills:
			d.title("Skills")
			for _, s := range cv.Skills {
				d.skill(s)
			}
		case model.SectionAchievements:
			d.title("Achievements")
			for _, a := range cv.Achievements {
				d.achievement(a)
			}
		}
	}

	return pdf.Output(w)
}

// harvardDoc carries the per-document layout state.
type harvardDoc struct {
	pdf              *fpdf.Fpdf
	o                Options
	nameSize         float64
	sectSize         float64
	dateSize         float64
	bodySize         float64
	lineH            float64
	indent           float64
	left             float64
	right            float64
	pageWidth        float64
	pageHeight       float64
	contentWidth     float64
	pageBreakTrigger float64
}

func ptToMM(pt float64) float64 { return pt / 72.0 * 25.4 }

func (d *harvardDoc) setFont(style string, size float64) {
	d.pdf.SetFont(fontFamily, style, size)
	d.pdf.SetTextColor(0, 0, 0)
}

// ensureSpace adds a page if the next h millimetres of content would overflow.
func (d *harvardDoc) ensureSpace(h float64) {
	if d.pdf.GetY()+h > d.pageBreakTrigger {
		d.pdf.AddPage()
	}
}

func (d *harvardDoc) header(b model.Basics) {
	d.setFont("B", d.nameSize)
	d.pdf.SetX(d.left)
	d.pdf.CellFormat(d.contentWidth, d.lineH*1.7, b.Name, "", 1, "C", false, 0, "")

	if strings.TrimSpace(b.Label) != "" {
		d.setFont("I", d.bodySize+1)
		d.pdf.CellFormat(d.contentWidth, d.lineH, b.Label, "", 1, "C", false, 0, "")
	}

	segs := d.contactSegments(b)
	if len(segs) > 0 {
		d.renderContact(segs)
	}
	d.pdf.Ln(d.lineH * 0.4)
}

// contactSegments builds the header contact line as linkable segments.
func (d *harvardDoc) contactSegments(b model.Basics) []segment {
	var parts []segment
	add := func(text, url string) {
		if strings.TrimSpace(text) != "" {
			parts = append(parts, segment{Text: text, URL: url})
		}
	}
	add(b.Phone, "")
	if b.Email != "" {
		add(b.Email, "mailto:"+b.Email)
	}
	add(b.Location, "")
	for _, p := range b.Profiles {
		add(p.Network, p.URL)
	}

	if len(parts) == 0 {
		return nil
	}
	sep := segment{Text: "  \u2022  "}
	var out []segment
	for i, p := range parts {
		if i > 0 {
			out = append(out, sep)
		}
		out = append(out, p)
	}
	return out
}

func (d *harvardDoc) renderContact(segs []segment) {
	d.setFont("", d.bodySize)

	total := 0.0
	for _, s := range segs {
		total += d.pdf.GetStringWidth(s.Text)
	}
	if total > d.contentWidth {
		// Too wide for one centered line: fall back to plain wrapped text.
		d.pdf.SetX(d.left)
		d.pdf.MultiCell(d.contentWidth, d.lineH, segmentText(segs), "", "C", false)
		return
	}
	d.pdf.SetX(d.left + (d.contentWidth-total)/2)
	for _, s := range segs {
		w := d.pdf.GetStringWidth(s.Text)
		d.pdf.CellFormat(w, d.lineH, s.Text, "", 0, "L", false, 0, s.URL)
	}
	d.pdf.Ln(d.lineH)
}

// title renders a section heading with a horizontal rule underneath.
func (d *harvardDoc) title(name string) {
	d.pdf.Ln(d.lineH * 0.55)
	d.ensureSpace(d.lineH * 2.2)
	d.setFont("B", d.sectSize)
	d.pdf.SetX(d.left)
	d.pdf.CellFormat(d.contentWidth, d.lineH*1.3, strings.ToUpper(name), "", 1, "L", false, 0, "")

	y := d.pdf.GetY()
	d.pdf.SetDrawColor(0, 0, 0)
	d.pdf.SetLineWidth(0.4)
	d.pdf.Line(d.left, y, d.right, y)
	d.pdf.SetY(y + d.lineH*0.5)
}

// entryLine draws a bold (optionally linked) title on the left and an italic
// date range aligned to the right. If both do not fit, the date wraps to the
// following line.
func (d *harvardDoc) entryLine(title string, date string, style string) {
	segs := splitSegments(title)

	dateWidth := 0.0
	if date != "" {
		d.setFont("I", d.dateSize)
		dateWidth = d.pdf.GetStringWidth(date)
	}

	d.setFont(style, d.bodySize)
	titleWidth := 0.0
	for _, s := range segs {
		titleWidth += d.pdf.GetStringWidth(s.Text)
	}

	const gap = 4.0
	d.ensureSpace(d.lineH)
	startY := d.pdf.GetY()

	if date != "" && titleWidth+gap+dateWidth <= d.contentWidth {
		d.pdf.SetXY(d.left, startY)
	} else if date != "" {
		// Title alone.
		d.pdf.SetXY(d.left, startY)
		for _, s := range segs {
			w := d.pdf.GetStringWidth(s.Text)
			d.pdf.CellFormat(w, d.lineH, s.Text, "", 0, "L", false, 0, s.URL)
		}
		d.pdf.Ln(d.lineH)
		d.ensureSpace(d.lineH)
		d.setFont("I", d.dateSize)
		d.pdf.SetX(d.left)
		d.pdf.CellFormat(d.contentWidth, d.lineH, date, "", 1, "R", false, 0, "")
		return
	}

	for _, s := range segs {
		w := d.pdf.GetStringWidth(s.Text)
		d.pdf.CellFormat(w, d.lineH, s.Text, "", 0, "L", false, 0, s.URL)
	}
	if date != "" {
		d.setFont("I", d.dateSize)
		d.pdf.SetX(d.right - dateWidth)
		d.pdf.SetY(startY)
		d.pdf.CellFormat(dateWidth, d.lineH, date, "", 0, "R", false, 0, "")
	}
	d.pdf.SetY(startY + d.lineH)
	d.pdf.SetX(d.left)
}

// subLine draws an italic secondary line (organization, location).
func (d *harvardDoc) subLine(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	d.setFont("I", d.bodySize)
	lines := d.pdf.SplitText(text, d.contentWidth)
	d.ensureSpace(float64(len(lines)) * d.lineH)
	d.pdf.SetX(d.left)
	d.pdf.MultiCell(d.contentWidth, d.lineH*0.95, text, "", "L", false)
}

// bullet renders a bullet point that wraps within the content width.
func (d *harvardDoc) bullet(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	d.setFont("", d.bodySize)
	lines := d.pdf.SplitText(text, d.contentWidth-d.indent)
	d.ensureSpace(float64(len(lines)) * d.lineH)
	d.pdf.SetX(d.left)
	d.pdf.CellFormat(d.indent, d.lineH, "\u2022", "", 0, "L", false, 0, "")
	d.pdf.MultiCell(d.contentWidth-d.indent, d.lineH, text, "", "L", false)
}

// paragraph renders a justified full-width paragraph.
func (d *harvardDoc) paragraph(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	d.setFont("", d.bodySize)
	lines := d.pdf.SplitText(text, d.contentWidth)
	d.ensureSpace(float64(len(lines)) * d.lineH)
	d.pdf.SetX(d.left)
	d.pdf.MultiCell(d.contentWidth, d.lineH, text, "", "J", false)
}

func (d *harvardDoc) education(e model.Education) {
	title := degreeTitle(e.Degree, e.Area)
	d.entryLine(title, model.FormatRange(e.Start, e.End), "B")

	sub := e.Institution
	if e.Location != "" {
		sub = joinComma(sub, e.Location)
	}
	d.subLine(sub)

	for _, h := range e.Highlights {
		d.bullet(h)
	}
	d.pdf.Ln(d.lineH * 0.35)
}

func (d *harvardDoc) experience(e model.Experience) {
	d.entryLine(e.Position, model.FormatRange(e.Start, e.End), "B")

	sub := e.Company
	if e.Location != "" {
		sub = joinComma(sub, e.Location)
	}
	d.subLine(sub)

	for _, h := range e.Highlights {
		d.bullet(h)
	}
	d.pdf.Ln(d.lineH * 0.35)
}

func (d *harvardDoc) project(p model.Project) {
	title := p.Name
	if p.URL != "" && !hasURL(splitSegments(p.Name)) {
		title = fmt.Sprintf("[%s](%s)", p.Name, p.URL)
	}
	d.entryLine(title, p.Date.FormatPoint(), "B")
	d.paragraph(p.Summary)
	for _, h := range p.Highlights {
		d.bullet(h)
	}
	d.pdf.Ln(d.lineH * 0.35)
}

func (d *harvardDoc) skill(s model.Skill) {
	d.setFont("B", d.bodySize)
	label := strings.TrimRight(s.Label, ":") + ": "
	labelWidth := d.pdf.GetStringWidth(label)

	detailsWidth := d.contentWidth - d.indent - labelWidth
	if detailsWidth < d.contentWidth*0.4 {
		// Long label: stack label and details on separate lines.
		d.ensureSpace(d.lineH)
		d.pdf.SetX(d.left)
		d.pdf.CellFormat(d.indent, d.lineH, "\u2022", "", 0, "L", false, 0, "")
		d.pdf.CellFormat(d.contentWidth-d.indent, d.lineH, strings.TrimSpace(label), "", 1, "L", false, 0, "")
		d.setFont("", d.bodySize)
		lines := d.pdf.SplitText(s.Details, d.contentWidth-d.indent)
		d.ensureSpace(float64(len(lines)) * d.lineH)
		d.pdf.SetX(d.left + d.indent)
		d.pdf.MultiCell(d.contentWidth-d.indent, d.lineH, s.Details, "", "L", false)
		return
	}

	lines := d.pdf.SplitText(s.Details, detailsWidth)
	d.ensureSpace(float64(len(lines)) * d.lineH)

	startY := d.pdf.GetY()
	d.pdf.SetXY(d.left, startY)
	d.pdf.CellFormat(d.indent, d.lineH, "\u2022", "", 0, "L", false, 0, "")
	d.pdf.CellFormat(labelWidth, d.lineH, label, "", 0, "L", false, 0, "")
	d.setFont("", d.bodySize)
	d.pdf.SetXY(d.left+d.indent+labelWidth, startY)
	d.pdf.MultiCell(detailsWidth, d.lineH, s.Details, "", "L", false)
}

func (d *harvardDoc) achievement(a model.Achievement) {
	d.entryLine(a.Name, a.Date.FormatPoint(), "B")
	for _, h := range a.Highlights {
		d.bullet(h)
	}
	d.pdf.Ln(d.lineH * 0.35)
}

// degreeTitle combines degree and area, e.g. "BS in Computer Science
// Engineering".
func degreeTitle(degree, area string) string {
	degree = strings.TrimSpace(degree)
	area = strings.TrimSpace(area)
	switch {
	case degree == "":
		return area
	case area == "":
		return degree
	default:
		return degree + " in " + area
	}
}

func joinComma(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + ", " + b
}
