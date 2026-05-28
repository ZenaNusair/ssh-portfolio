package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(180*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type particle struct {
	x, y int
	ch   rune
}

const (
	snowHeight      = 5 // total rows in the snow + name overlay
	figletStartRow  = 1 // figlet "zena" occupies rows 1..3 inside the overlay
	snowW           = 60
)

var snowChars = []rune{'+', '·', '.', '*'}

type tab int

const (
	tabAbout tab = iota
	tabExperience
	tabSkills
	tabEducation
	tabContact
)

var tabNames = []string{"about", "experience", "skills", "education", "contact"}

const nameArt = ` ___ ___ ___  ___ _
/_ // -_) _ \/ _ ` + "`" + `/
/__/\__/_//_/\_,_/`

type Model struct {
	width      int
	height     int
	current    tab
	particles  []particle
	blink      bool   // blinking cursor toggle
	frame      int    // tick counter
	showReflection bool   // press `?` to reveal the reflection essay
	fortune    string // rotating one-liner shown at the bottom
}

var fortunes = []string{
	"today's vibe: ship it before the meeting.",
	"\"works on my machine\" — every dev, eventually.",
	"semicolon optional, regret mandatory.",
	"git push --force is a personality.",
	"the bug was in the last place I looked. it always is.",
	"stack overflow is just a study group, your honor.",
	"my code has a 50% success rate; it works in production OR locally.",
	"the cloud is just someone else's terminal.",
	"reading the docs is a personality flaw I'm working on.",
	"caffeine: the original package manager.",
}

func NewModel(w, h int) Model {
	return Model{
		width:   w,
		height:  h,
		current: tabAbout,
		fortune: fortunes[rand.Intn(len(fortunes))],
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			m.showReflection = false
			if m.current < tabContact {
				m.current++
			}
			m.fortune = fortunes[rand.Intn(len(fortunes))]
		case "shift+tab", "left", "h":
			m.showReflection = false
			if m.current > tabAbout {
				m.current--
			}
			m.fortune = fortunes[rand.Intn(len(fortunes))]
		case "?":
			m.showReflection = !m.showReflection
		case "f":
			m.fortune = fortunes[rand.Intn(len(fortunes))]
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.frame++
		// toggle blinking cursor every ~3 ticks (~540ms)
		if m.frame%3 == 0 {
			m.blink = !m.blink
		}
		next := m.particles[:0]
		for _, p := range m.particles {
			p.y++
			if p.y < snowHeight {
				next = append(next, p)
			}
		}
		// spawn at most 1 particle per tick — keep it sparse
		if rand.Intn(2) == 0 {
			next = append(next, particle{
				x:  rand.Intn(snowW),
				y:  0,
				ch: snowChars[rand.Intn(len(snowChars))],
			})
		}
		m.particles = next
		return m, tickCmd()
	}
	return m, nil
}

// renderOverlay draws the figlet name (centered) with snow particles falling
// through it. Particles overlay the figlet chars where they collide.
func renderOverlay(particles []particle, accent lipgloss.Color, w int) string {
	figletLines := strings.Split(nameArt, "\n")

	// center the figlet horizontally inside w
	figW := 0
	for _, l := range figletLines {
		if len(l) > figW {
			figW = len(l)
		}
	}
	figCol := (w - figW) / 2
	if figCol < 0 {
		figCol = 0
	}

	// quick lookup: which cell has a particle and which char
	particleAt := make(map[[2]int]rune, len(particles))
	for _, p := range particles {
		particleAt[[2]int{p.x, p.y}] = p.ch
	}

	accentStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)

	var out strings.Builder
	for row := 0; row < snowHeight; row++ {
		var figRow string
		if row >= figletStartRow && row-figletStartRow < len(figletLines) {
			figRow = figletLines[row-figletStartRow]
		}
		for col := 0; col < w; col++ {
			if ch, ok := particleAt[[2]int{col, row}]; ok {
				out.WriteString(styleMuted.Render(string(ch)))
				continue
			}
			figIdx := col - figCol
			if figIdx >= 0 && figIdx < len(figRow) && figRow[figIdx] != ' ' {
				out.WriteString(accentStyle.Render(string(figRow[figIdx])))
				continue
			}
			out.WriteByte(' ')
		}
		if row < snowHeight-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func (m Model) View() string {
	if m.width < 100 {
		return "terminal too narrow — please widen to at least 100 cols\n"
	}

	leftW := 52 // portrait is 50 cols + 2 padding
	gap := 4
	rightW := m.width - leftW - gap
	if rightW < 59 {
		rightW = 59 // tab bar needs ~59 cols
	}

	portrait := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Width(leftW).
		Render(Portrait)

	// ── Right column content ──────────────────────────────────
	// figlet name with snow falling through it (active tab's accent)
	accent := tabColors[m.current]
	nameWithSnow := renderOverlay(m.particles, accent, snowW)

	// section content (changes per tab)
	var content string
	if m.showReflection {
		content = m.viewReflection()
	} else {
		switch m.current {
		case tabAbout:
			content = m.viewAbout()
		case tabExperience:
			content = m.viewExperience()
		case tabSkills:
			content = m.viewSkills()
		case tabEducation:
			content = m.viewEducation()
		case tabContact:
			content = m.viewContact()
		}
	}

	// horizontal tab bar — each tab carries its own loud color, active is bold + ✦
	var tabBar strings.Builder
	for i, n := range tabNames {
		if i > 0 {
			tabBar.WriteString("   ")
		}
		c := tabColors[i]
		if tab(i) == m.current {
			tabBar.WriteString(lipgloss.NewStyle().Foreground(c).Bold(true).Render("✦ " + n))
		} else {
			tabBar.WriteString(lipgloss.NewStyle().Foreground(c).Faint(true).Render("  " + n))
		}
	}

	help := styleHelp.Render("[← → switch · q quit]")
	fortuneLine := lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		Render("» " + m.fortune)

	// mori-style stack: name-with-snow → content → pad → tab bar (at bottom)
	top := nameWithSnow + "\n\n" + content

	portraitHeight := lipgloss.Height(portrait)
	topHeight := lipgloss.Height(top)
	tabHeight := lipgloss.Height(tabBar.String())
	pad := portraitHeight - topHeight - tabHeight
	if pad < 1 {
		pad = 1
	}
	right := top + strings.Repeat("\n", pad) + tabBar.String()

	rightCol := lipgloss.NewStyle().Width(rightW).Render(right)

	body := lipgloss.JoinHorizontal(lipgloss.Top, portrait, strings.Repeat(" ", gap), rightCol)

	// help + rotating fortune sit at the very bottom under the portrait
	full := body + "\n" + help + "\n" + fortuneLine

	return styleBase.
		Width(m.width).
		Height(m.height).
		Render(full)
}

// ── Section views ─────────────────────────────────────────────

func (m Model) cursor() string {
	if m.blink {
		return lipgloss.NewStyle().Foreground(tabColors[m.current]).Render("▮")
	}
	return " "
}

func (m Model) sectionHeader(label string) string {
	accent := tabColors[m.current]
	return lipgloss.NewStyle().Foreground(accent).Bold(true).Render("~ "+label) + m.cursor()
}

func (m Model) viewAbout() string {
	accent := tabColors[m.current]
	primary := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	loud := lipgloss.NewStyle().Foreground(accent).Bold(true)
	return fmt.Sprintf(`%s

%s

%s

%s
`,
		m.sectionHeader("whoami"),
		primary.Render(
			"a software engineer who builds react and\n"+
				"react native apps that hold up when real\n"+
				"users show up."),
		styleMuted.Render(
			"she likes clean architecture, fast sql, dashboards\n"+
				"that tell the truth, and slipping AI into places\n"+
				"it actually earns its keep.\n\n"+
				"her commit messages are 30% emoji, 70% 'fix' —\n"+
				"a problem for future zena, not present zena."),
		loud.Render("> press ? for my reflection on AI, creativity, and human purpose"),
	)
}

func (m Model) viewExperience() string {
	var b strings.Builder
	accent := tabColors[m.current]
	dot := lipgloss.NewStyle().Foreground(accent).Render("·")
	arrow := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(">")

	b.WriteString(m.sectionHeader("work.log") + "\n\n")

	b.WriteString(arrow + " " + styleEntryTitle.Render("software engineer") +
		styleMuted.Render(" @ firnas") + "\n")
	b.WriteString(styleEntryMeta.Render("  may 2024 → jan 2025") + "\n\n")
	bullets := []string{
		"shipped react + react native apps that didn't immediately catch fire",
		"wired role-based auth before 'vibes-based auth' became a thing",
		"made queries faster — some users noticed",
		"built dashboards. so many dashboards.",
	}
	for _, bt := range bullets {
		b.WriteString("  " + dot + " " + styleMuted.Render(bt) + "\n")
	}

	b.WriteString("\n")

	b.WriteString(arrow + " " + styleEntryTitle.Render("power apps developer") +
		styleMuted.Render(" @ power apps technology") + "\n")
	b.WriteString(styleEntryMeta.Render("  feb 2023 → mar 2024") + "\n\n")
	bullets2 := []string{
		"power bi dashboards (they powered things, true to the name)",
		"automated workflows so humans could nap",
		"low-code solutions with high-effort polish",
	}
	for _, bt := range bullets2 {
		b.WriteString("  " + dot + " " + styleMuted.Render(bt) + "\n")
	}

	return b.String()
}

func (m Model) viewSkills() string {
	var b strings.Builder
	accent := tabColors[m.current]
	dot := lipgloss.NewStyle().Foreground(accent).Render("·")

	b.WriteString(m.sectionHeader("inventory") + "\n\n")
	b.WriteString(styleMuted.Italic(true).Render("(opens backpack)") + "\n\n")

	b.WriteString(styleEntryTitle.Render("languages spoken to computers") + "\n\n")
	langs := []string{"TypeScript", "JavaScript", "Python", "Go", "Java", "Rust", "C++", "SQL"}
	b.WriteString(gridRow(langs, 4, 14, dot))

	b.WriteString("\n")

	b.WriteString(styleEntryTitle.Render("things i actually use") + "\n\n")
	tools := []string{"React", "React Native", "REST APIs", "CI/CD", "Power BI", "Power Automate", "AI integration", "End-to-End", "Deployments"}
	b.WriteString(gridRow(tools, 4, 18, dot))

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(accent).Render(
		"xp unlocked: lots.  boss fights: ongoing."))

	return b.String()
}

// gridRow lays items out in a grid with `cols` columns of width `colW`.
func gridRow(items []string, cols, colW int, dot string) string {
	var b strings.Builder
	for i, s := range items {
		cell := "· " + s
		b.WriteString("  " + lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(colW).
			Render(strings.Replace(cell, "·", dot, 1)))
		if (i+1)%cols == 0 {
			b.WriteString("\n")
		}
	}
	if len(items)%cols != 0 {
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) viewEducation() string {
	var b strings.Builder
	accent := tabColors[m.current]
	dot := lipgloss.NewStyle().Foreground(accent).Render("·")
	check := lipgloss.NewStyle().Foreground(accent).Bold(true).Render("✓")

	b.WriteString(m.sectionHeader("school.txt") + "\n\n")

	b.WriteString(styleEntryTitle.Render("B.Sc. Computer Science") + "\n")
	b.WriteString(styleMuted.Render("Applied Science Private University · Amman") + "\n\n")

	quest := []string{
		"took finals",
		"passed finals",
		"earned degree",
		"unlocked: real job (terms apply)",
	}
	for _, q := range quest {
		b.WriteString("  " + check + " " + styleMuted.Render(q) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(styleEntryTitle.Render("receipts") + "\n\n")
	certs := []string{
		"Cisco — Data Analytics Essentials",
		"Correlation One — Tech for Jobs PDP",
	}
	for _, c := range certs {
		b.WriteString("  " + dot + " " + styleMuted.Render(c) + "\n")
	}

	return b.String()
}

func (m Model) viewContact() string {
	accent := tabColors[m.current]
	label := lipgloss.NewStyle().Foreground(accent).Bold(true).Render
	return fmt.Sprintf(`%s

%s

  %s   zena.m.nusair@gmail.com
  %s   +962 791 759 868
  %s   github.com/ZenaNusair
  %s   ssh ssh.zena.dev

%s
`,
		m.sectionHeader("ping me"),
		styleMuted.Render("say hi. talk shop. argue about tabs vs spaces.\ni respond faster than the average api."),
		label(" @"),
		label(" #"),
		label(" ⌥"),
		label(" →"),
		styleMuted.Italic(true).Render("("+`pro tip: ssh is the fastest way. you're already here.`+")"),
	)
}

func (m Model) viewReflection() string {
	accent := tabColors[m.current]
	loud := lipgloss.NewStyle().Foreground(accent).Bold(true)
	return fmt.Sprintf(`%s

%s

%s

%s

%s

%s
`,
		m.sectionHeader("reflection.md"),
		loud.Render("AI, Creativity, and Human Purpose"),
		styleMuted.Render(
			"AI is often viewed with skepticism in creative spaces,\n"+
				"yet I see it less as a replacement for human ingenuity\n"+
				"and more as a tool that can expand it."),
		styleMuted.Render(
			"Rather than eliminating creativity or human labor, AI\n"+
				"has the potential to reduce cognitive barriers, inspire\n"+
				"new ideas, and shift human contribution toward more\n"+
				"meaningful, higher-order work."),
		styleMuted.Render(
			"Drawing from both personal experience in computer science\n"+
				"and broader discussions in philosophy, HCI, and economics,\n"+
				"I argue that the real challenge AI presents is not the\n"+
				"disappearance of human value, but the need to redefine\n"+
				"purpose, creativity, and collaboration in an increasingly\n"+
				"automated world."),
		styleMuted.Italic(true).Render("— zena · press ? to return"),
	)
}
