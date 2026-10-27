package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/miguehm/yml-to-cv/assets/fonts"
	assetIcons "github.com/miguehm/yml-to-cv/assets/icons"
	"github.com/miguehm/yml-to-cv/internal/i18n"
	"github.com/miguehm/yml-to-cv/internal/icons"
	"github.com/miguehm/yml-to-cv/internal/model"
)

const (
	fontFamily     = "Tinos"
	iconFontFamily = "Icons"
	iconSizeRatio  = 0.85
	iconGapMM      = 0.9
	// iconCenterRatio is the vertical centre of a Bootstrap Icons glyph above
	// the baseline, expressed as a fraction of the icon em. textCapRatio is the
	// cap height of Tinos, used to centre icons on the body text optically.
	iconCenterRatio  = 0.51
	textCapRatio     = 0.662
	contactSeparator = "  \u2022  "
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
	if o.Language == "" {
		o.Language = i18n.English
	}
	return o
}

func (h *harvardRenderer) Render(w io.Writer, cv *model.CV) error {
	o := h.opts.withDefaults()

	pdf := fpdf.New("P", "mm", strings.ToLower(o.Paper), "")
	pdf.SetMargins(o.Margin, o.Margin, o.Margin)
	pdf.SetAutoPageBreak(true, o.Margin)
	pdf.SetCompression(true)
	pdf.SetCatalogSort(true)
	pdf.SetCreationDate(o.CreationDate)
	pdf.SetModificationDate(o.CreationDate)
	pdf.SetTitle(cv.Basics.Name+" - CV", true)
	pdf.SetAuthor(cv.Basics.Name, true)

	pdf.AddUTF8FontFromBytes(fontFamily, "", fonts.TinosRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", fonts.TinosBold)
	pdf.AddUTF8FontFromBytes(fontFamily, "I", fonts.TinosItalic)
	pdf.AddUTF8FontFromBytes(fontFamily, "BI", fonts.TinosBoldItalic)
	pdf.AddUTF8FontFromBytes(iconFontFamily, "", assetIcons.BootstrapIcons)

	pdf.AddPage()

	d := &harvardDoc{
		pdf:       pdf,
		o:         o,
		nameSize:  defaultNameSize,
		sectSize:  defaultSectSize,
		dateSize:  defaultDateSize,
		bodySize:  o.BodySize,
		iconSize:  o.BodySize * iconSizeRatio,
		lineH:     ptToMM(o.BodySize) * 1.28,
		indent:    4.0,
		left:      o.Margin,
		lang:      o.Language,
		cat:       i18n.For(o.Language),
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
			d.title(d.cat.Education)
			for _, e := range cv.Education {
				d.education(e)
			}
		case model.SectionExperience:
			d.title(d.cat.Experience)
			for _, e := range cv.Experience {
				d.experience(e)
			}
		case model.SectionProjects:
			d.title(d.cat.Projects)
			for _, p := range cv.Projects {
				d.project(p)
			}
		case model.SectionSkills:
			d.title(d.cat.Skills)
			for _, s := range cv.Skills {
				d.skill(s)
			}
		case model.SectionAchievements:
			d.title(d.cat.Achievements)
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
	iconSize         float64
	lineH            float64
	indent           float64
	left             float64
	right            float64
	pageWidth        float64
	pageHeight       float64
	contentWidth     float64
	pageBreakTrigger float64
	lang             i18n.Language
	cat              i18n.Catalog
}

func ptToMM(pt float64) float64 { return pt / 72.0 * 25.4 }

func (d *harvardDoc) setFont(style string, size float64) {
	d.pdf.SetFont(fontFamily, style, size)
	d.pdf.SetTextColor(0, 0, 0)
}

// setIconFont selects the icon typeface at the configured icon size. Icons
// share the text colour so they match the classic monochrome Harvard header.
func (d *harvardDoc) setIconFont() {
	d.pdf.SetFont(iconFontFamily, "", d.iconSize)
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

	d.renderContact(d.contactItems(b))

	if strings.TrimSpace(b.Summary) != "" {
		d.pdf.Ln(d.lineH * 0.4)
		d.paragraph(b.Summary)
	}

	d.pdf.Ln(d.lineH * 0.4)
}

// contactItem is a single header contact detail, optionally prefixed by an
// icon glyph, and optionally linked to a URL.
type contactItem struct {
	Icon string
	Text string
	URL  string
}

// contactItems builds the header contact details in display order. Icons can be
// turned off entirely with "basics.icons: false" and overridden per profile
// with "icon:".
func (d *harvardDoc) contactItems(b model.Basics) []contactItem {
	showIcons := b.Icons == nil || *b.Icons

	var items []contactItem
	add := func(icon, text, url string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		if !showIcons {
			icon = ""
		}
		items = append(items, contactItem{Icon: icon, Text: text, URL: url})
	}
	add(icons.Telephone, b.Phone, "")
	if b.Email != "" {
		add(icons.Envelope, b.Email, "mailto:"+b.Email)
	}
	add(icons.Location, b.Location, "")
	for _, p := range b.Profiles {
		icon := icons.ForNetwork(p.Network)
		if p.Icon != "" {
			icon = p.Icon
		}
		// Show the shortened URL so the address is readable when printed; the
		// full URL stays as the clickable target.
		text := p.Network
		if short := shortURL(p.URL); short != "" {
			text = short
		}
		add(icon, text, p.URL)
	}
	return items
}

// contactItemWidth measures an item including its icon and the gap that
// separates the glyph from the text.
func (d *harvardDoc) contactItemWidth(it contactItem) float64 {
	w := 0.0
	if it.Icon != "" {
		d.setIconFont()
		w += d.pdf.GetStringWidth(string(icons.Rune(it.Icon))) + iconGapMM
	}
	d.setFont("", d.bodySize)
	return w + d.pdf.GetStringWidth(it.Text)
}

// drawContactItem draws one item at (x, y) advancing the icon/text baseline so
// both sit on the same line.
func (d *harvardDoc) drawContactItem(x, y float64, it contactItem) {
	if it.Icon != "" {
		d.setIconFont()
		glyph := string(icons.Rune(it.Icon))
		gw := d.pdf.GetStringWidth(glyph)
		// fpdf places the baseline at y+0.5*h+0.3*size (size in page units).
		// Correct for the smaller icon, then drop it a hair so the glyph centre
		// meets the cap-height centre of the body text.
		nudge := ptToMM(0.3*(d.bodySize-d.iconSize)) + d.iconVerticalOffset()
		d.pdf.SetXY(x, y+nudge)
		d.pdf.CellFormat(gw, d.lineH, glyph, "", 0, "L", false, 0, it.URL)
		x += gw + iconGapMM
	}
	d.setFont("", d.bodySize)
	tw := d.pdf.GetStringWidth(it.Text)
	d.pdf.SetXY(x, y)
	d.pdf.CellFormat(tw, d.lineH, it.Text, "", 0, "L", false, 0, it.URL)
}

// iconVerticalOffset is the extra downward shift, in millimetres, that lines up
// an icon's glyph centre with the cap-height centre of the body text.
func (d *harvardDoc) iconVerticalOffset() float64 {
	return iconCenterRatio*ptToMM(d.iconSize) - 0.5*textCapRatio*ptToMM(d.bodySize)
}

// renderContact draws the contact details centered on one or more lines,
// separated by bullets, wrapping only when the content does not fit.
func (d *harvardDoc) renderContact(items []contactItem) {
	if len(items) == 0 {
		return
	}

	widths := make([]float64, len(items))
	for i, it := range items {
		widths[i] = d.contactItemWidth(it)
	}
	d.setFont("", d.bodySize)
	sepWidth := d.pdf.GetStringWidth(contactSeparator)

	var lines [][]int
	var cur []int
	curWidth := 0.0
	for i := range items {
		if len(cur) == 0 {
			cur, curWidth = []int{i}, widths[i]
			continue
		}
		if curWidth+sepWidth+widths[i] <= d.contentWidth {
			cur = append(cur, i)
			curWidth += sepWidth + widths[i]
			continue
		}
		lines = append(lines, cur)
		cur, curWidth = []int{i}, widths[i]
	}
	lines = append(lines, cur)

	for _, line := range lines {
		total := 0.0
		for j, idx := range line {
			if j > 0 {
				total += sepWidth
			}
			total += widths[idx]
		}
		d.ensureSpace(d.lineH)
		y := d.pdf.GetY()
		x := d.left + (d.contentWidth-total)/2
		if x < d.left {
			x = d.left
		}
		for j, idx := range line {
			if j > 0 {
				d.setFont("", d.bodySize)
				d.pdf.SetXY(x, y)
				d.pdf.CellFormat(sepWidth, d.lineH, contactSeparator, "", 0, "L", false, 0, "")
				x += sepWidth
			}
			d.drawContactItem(x, y, items[idx])
			x += widths[idx]
		}
		d.pdf.SetY(y + d.lineH)
		d.pdf.SetX(d.left)
	}
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
		d.pdf.SetXY(d.right-dateWidth, startY)
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
	title := degreeTitle(e.Degree, e.Area, d.cat.DegreeConnector)
	d.entryLine(title, model.FormatRangeIn(e.Start, e.End, d.lang), "B")

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
	d.entryLine(e.Position, model.FormatRangeIn(e.Start, e.End, d.lang), "B")

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
		// Print the shortened URL next to the name so the address is visible on
		// paper, while the name and the URL both link to the full URL on screen.
		title = fmt.Sprintf("[%s](%s) ([%s](%s))", p.Name, p.URL, shortURL(p.URL), p.URL)
	}
	d.entryLine(title, p.Date.FormatPointIn(d.lang), "B")
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
	d.entryLine(a.Name, a.Date.FormatPointIn(d.lang), "B")
	for _, h := range a.Highlights {
		d.bullet(h)
	}
	d.pdf.Ln(d.lineH * 0.35)
}

// degreeTitle combines degree and area with the connector of the active
// language, e.g. "BS in Computer Science" or "BS en Computer Science".
func degreeTitle(degree, area, connector string) string {
	degree = strings.TrimSpace(degree)
	area = strings.TrimSpace(area)
	if connector = strings.TrimSpace(connector); connector != "" {
		connector += " "
	}
	switch {
	case degree == "":
		return area
	case area == "":
		return degree
	default:
		return degree + " " + connector + area
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
