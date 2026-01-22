package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wb "github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

// --- 1. DATA & CONTENT ---

type Item struct {
	Title       string
	Category    string
	Description string
	TechStack   string
	Tag         string
	Icon        string
	Link        string
}

var items = []Item{
	// PROJECTS
	{
		Title:     "zenRoute",
		Category:  "projects",
		Tag:       "IoT",
		Icon:      "◈",
		Link:      "",
		TechStack: "FastAPI · Supabase · Docker · React",
		Description: "Leading a 6-member team to build Sri Lanka's first smart transport safety platform. " +
			"Features real-time IoT hardware integration, ML-based ETA predictions, and a scalable backend architecture.",
	},
	{
		Title:     "PathHelm",
		Category:  "projects",
		Tag:       "API",
		Icon:      "⬡",
		Link:      "https://github.com/KingSajxxd/pathhelm",
		TechStack: "Python · Redis · Docker · AI",
		Description: "A developer-first, containerized API gateway built for speed. " +
			"Handles rate limiting, API key validation, and logs traffic. Includes AI-powered traffic analytics.",
	},
	{
		Title:     "Chat Server",
		Category:  "projects",
		Tag:       "Backend",
		Icon:      "◉",
		Link:      "https://github.com/KingSajxxd/python-chat-server",
		TechStack: "Python · WebSockets · Docker",
		Description: "Real-Time WebSocket Backend. A containerized, event-driven chat server built for async communication. " +
			"Uses a custom WebSocket ConnectionManager and circular buffer to broadcast messages.",
	},

	// ABOUT
	{
		Title:     "About",
		Category:  "about",
		Tag:       "Profile",
		Icon:      "●",
		TechStack: "Backend · DevOps · Cloud",
		Description: "Software Engineering Undergraduate at IIT/Westminster. " +
			"Focused on Backend Development and Cloud Solutions. " +
			"I don't just write code; I ship systems.",
	},
}

// Social links with actual URLs
type Social struct {
	Icon string
	Name string
	URL  string
	Link string
}

var socials = []Social{
	{Icon: "◐ ", Name: "GitHub", URL: "github.com/KingSajxxd", Link: "https://github.com/KingSajxxd"},
	{Icon: "◧ ", Name: "LinkedIn", URL: "linkedin.com/in/sajjad-aiyoob", Link: "https://linkedin.com/in/sajjad-aiyoob"},
	{Icon: "✉ ", Name: "Email", URL: "sajaiyoobofficial@gmail.com", Link: "mailto:sajaiyoobofficial@gmail.com"},
}

// Easter egg quotes
var quotes = []string{
	"\"First, solve the problem. Then, write the code.\"",
	"\"Code is like humor. When you have to explain it, it's bad.\"",
	"\"Any fool can write code that a computer can understand.\"",
	"\"Talk is cheap. Show me the code.\" - Linus Torvalds",
	"\"It works on my machine.\" - Every developer ever",
	"\"There are only 10 types of people in the world...\"",
	"\"rm -rf / # Trust me, it works\" - Nobody ever",
	"\"sudo make me a sandwich\" - xkcd",
}

// Easter egg secrets
var easterEggHints = []string{
	"Try typing 'hello'...",
	"Press 's' for a surprise...",
	"The konami code works here...",
	"Press 'c' for confetti...",
}

// --- 2. STYLES ---

var (
	// Colors - Minimal palette
	accent  = lipgloss.Color("#FF6B35")
	accent2 = lipgloss.Color("#FF8C42")
	subtle  = lipgloss.Color("#4A4A4A")
	muted   = lipgloss.Color("#666666")
	dimmed  = lipgloss.Color("#3A3A3A")
	fg      = lipgloss.Color("#FAFAFA")
	fgDim   = lipgloss.Color("#888888")
	green   = lipgloss.Color("#00FF88")
	cyan    = lipgloss.Color("#00D4FF")

	// Logo style
	logoStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	// Navigation
	navActive = lipgloss.NewStyle().
			Foreground(fg).
			Background(dimmed).
			Padding(0, 2).
			Bold(true)

	navInactive = lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 2)

	// Section header
	sectionStyle = lipgloss.NewStyle().
			Foreground(muted).
			Bold(true).
			MarginTop(1).
			MarginBottom(0)

	// Item styles
	itemNormal = lipgloss.NewStyle().
			Foreground(fgDim).
			PaddingLeft(2)

	itemSelected = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			PaddingLeft(1)

	tagStyle = lipgloss.NewStyle().
			Foreground(muted).
			Background(dimmed).
			Padding(0, 1)

	// Detail view
	detailBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimmed).
			Padding(1, 2).
			Width(56)

	titleStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	techStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Italic(true)

	descStyle = lipgloss.NewStyle().
			Foreground(fg).
			Width(50)

	socialIcon = lipgloss.NewStyle().
			Foreground(accent)

	socialText = lipgloss.NewStyle().
			Foreground(fgDim)

	hintStyle = lipgloss.NewStyle().
			Foreground(muted).
			Italic(true)

	// Splash styles
	splashStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	// Quote style
	quoteStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Italic(true)
)

// --- 3. MODEL ---

const (
	ViewSplash = iota
	ViewList
	ViewDetail
	ViewMatrix // Easter egg
	ViewHelp   // New help view
)

// Messages for animations
type tickMsg time.Time
type blinkMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func blinkCmd() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
		return blinkMsg{}
	})
}

type model struct {
	cursor int
	view   int
	width  int
	height int
	// Splash animation
	splashText  string
	splashIndex int
	blinkCount  int
	showCursor  bool
	splashDone  bool
	// Easter eggs
	konamiIndex    int
	matrixRain     [][]rune
	matrixTick     int
	showQuote      bool
	currentQuote   string
	typedBuffer    string
	showConfetti   bool
	confettiTick   int
	showSnake      bool
	snakeX         int
	snakeY         int
	snakeDir       int
	showHint       bool
	hintIndex      int
	easterEggTimer int
}

var splashFullText = "Initializing portfolio...\n> Loading projects\n> Connecting systems\n> Welcome, visitor."
var konamiCode = []string{"up", "up", "down", "down", "left", "right", "left", "right", "b", "a"}

func initialModel() model {
	return model{
		cursor:      0,
		view:        ViewSplash,
		splashText:  "",
		splashIndex: 0,
		blinkCount:  0,
		showCursor:  true,
		konamiIndex: 0,
		typedBuffer: "",
		snakeX:      10,
		snakeY:      5,
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

// --- 4. UPDATE ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Initialize matrix rain
		if m.matrixRain == nil {
			m.matrixRain = make([][]rune, msg.Width)
			for i := range m.matrixRain {
				m.matrixRain[i] = make([]rune, msg.Height)
			}
		}

	case tickMsg:
		if m.view == ViewSplash && !m.splashDone {
			if m.splashIndex < len(splashFullText) {
				m.splashText += string(splashFullText[m.splashIndex])
				m.splashIndex++
				return m, tickCmd()
			} else {
				m.splashDone = true
				return m, blinkCmd()
			}
		}
		if m.view == ViewMatrix {
			m.matrixTick++
			// Update matrix rain
			for i := range m.matrixRain {
				if rand.Intn(10) < 3 {
					for j := len(m.matrixRain[i]) - 1; j > 0; j-- {
						m.matrixRain[i][j] = m.matrixRain[i][j-1]
					}
					chars := "アイウエオカキクケコサシスセソタチツテト0123456789ABCDEF"
					m.matrixRain[i][0] = rune(chars[rand.Intn(len(chars))])
				}
			}
			return m, tickCmd()
		}

		// Auto-hide easter eggs after 10 seconds
		if m.view == ViewList && (m.showQuote || m.showConfetti) {
			m.easterEggTimer++
			if m.easterEggTimer > 200 { // 10 seconds (200 * 50ms)
				m.showQuote = false
				m.showConfetti = false
				m.easterEggTimer = 0
				return m, nil
			}
			return m, tickCmd()
		}

	case blinkMsg:
		if m.view == ViewSplash && m.splashDone {
			m.showCursor = !m.showCursor
			m.blinkCount++
			if m.blinkCount >= 6 { // 3 full blinks
				m.view = ViewList
				return m, nil
			}
			return m, blinkCmd()
		}

	case tea.KeyMsg:
		key := msg.String()

		// Skip splash on any key
		if m.view == ViewSplash {
			m.view = ViewList
			return m, nil
		}

		// Exit matrix mode
		if m.view == ViewMatrix {
			if key == "esc" || key == "q" {
				m.view = ViewList
			}
			return m, tickCmd()
		}

		// Konami code detection
		if key == konamiCode[m.konamiIndex] {
			m.konamiIndex++
			if m.konamiIndex == len(konamiCode) {
				m.konamiIndex = 0
				m.view = ViewMatrix
				return m, tickCmd()
			}
		} else {
			m.konamiIndex = 0
		}

		// Track typed characters for easter eggs
		if len(key) == 1 {
			m.typedBuffer += key
			if len(m.typedBuffer) > 10 {
				m.typedBuffer = m.typedBuffer[1:]
			}
			// Check for secret words
			if strings.HasSuffix(m.typedBuffer, "hello") {
				m.showQuote = true
				m.currentQuote = "👋 Hello there, curious one! You found a secret!"
				m.typedBuffer = ""
				m.easterEggTimer = 0
				return m, tickCmd()
			}
			if strings.HasSuffix(m.typedBuffer, "hire") {
				m.showQuote = true
				m.currentQuote = "💼 I'm available! Email: sajaiyoobofficial@gmail.com"
				m.typedBuffer = ""
				m.easterEggTimer = 0
				return m, tickCmd()
			}
		}

		switch key {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.view == ViewList {
				return m, tea.Quit
			}
			m.view = ViewList

		case "up", "k":
			if m.view == ViewList && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.view == ViewList && m.cursor < len(items)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.view == ViewList {
				m.view = ViewDetail
			}

		case "esc", "backspace":
			if m.view == ViewDetail || m.view == ViewHelp {
				m.view = ViewList
			}
			m.showQuote = false
			m.showConfetti = false
			m.showHint = false

		case "?":
			// Toggle help view
			if m.view == ViewHelp {
				m.view = ViewList
			} else {
				m.view = ViewHelp
			}

		case "s":
			// Surprise easter egg
			m.showQuote = true
			m.currentQuote = "🐍 Ssssurprise! You found me!"
			m.showConfetti = true
			m.confettiTick = 0
			m.easterEggTimer = 0
			return m, tickCmd()

		case "c":
			// Confetti easter egg
			m.showConfetti = !m.showConfetti
			if m.showConfetti {
				m.confettiTick = 0
				m.easterEggTimer = 0
				return m, tickCmd()
			}

		case "m":
			// Matrix mode shortcut (easier than konami)
			m.view = ViewMatrix
			return m, tickCmd()

		case "tab":
			// Cycle through items faster
			m.cursor = (m.cursor + 1) % len(items)
		}
	}
	return m, nil
}

// --- 5. VIEW ---

// OSC 8 hyperlink helper
func hyperlink(url, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

func centerText(s string, width int) string {
	lines := strings.Split(s, "\n")
	var centered []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth >= width {
			centered = append(centered, line)
			continue
		}
		padding := (width - lineWidth) / 2
		centered = append(centered, strings.Repeat(" ", padding)+line)
	}
	return strings.Join(centered, "\n")
}

func (m model) renderSplash() string {
	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	var b strings.Builder

	// Terminal-style splash
	b.WriteString("\n\n")

	// FIX: Remove indentation so the box aligns perfectly
	asciiTerminal := `┌───────────────────────────────┐
│  ▓▓▓  SAJJAD'S TERMINAL  ▓▓▓  │
└───────────────────────────────┘`
	b.WriteString(centerText(logoStyle.Render(asciiTerminal), width))
	b.WriteString("\n\n")

	// Typing effect text
	displayText := splashStyle.Render(m.splashText)
	cursor := ""
	if m.showCursor {
		cursor = cursorStyle.Render("█")
	} else {
		cursor = " "
	}
	b.WriteString(centerText(displayText+cursor, width))

	// Center vertically
	content := b.String()
	contentHeight := strings.Count(content, "\n")
	topPadding := (height - contentHeight) / 3

	return strings.Repeat("\n", topPadding) + content
}

func (m model) renderMatrix() string {
	var b strings.Builder
	for y := 0; y < m.height-2; y++ {
		for x := 0; x < m.width; x++ {
			if x < len(m.matrixRain) && y < len(m.matrixRain[x]) && m.matrixRain[x][y] != 0 {
				intensity := 255 - (y * 10)
				if intensity < 50 {
					intensity = 50
				}
				color := lipgloss.Color(fmt.Sprintf("#00%02x00", intensity))
				b.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(m.matrixRain[x][y])))
			} else {
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}
	hint := centerText(hintStyle.Render("You found the Matrix! Press ESC to return..."), m.width)
	b.WriteString(hint)
	return b.String()
}

func (m model) View() string {
	// Splash screen
	if m.view == ViewSplash {
		return m.renderSplash()
	}

	// Matrix easter egg
	if m.view == ViewMatrix {
		return m.renderMatrix()
	}

	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// Content width logic - adapt to smaller screens
	var contentWidth int

	if width < 50 {
		// Tight layout for mobile/small terms
		contentWidth = width - 4
	} else {
		// Spacious layout for desktop
		contentWidth = width - 10
		if contentWidth > 72 {
			contentWidth = 72
		}
	}

	if contentWidth < 30 {
		contentWidth = 30 // Absolute minimum to prevents rendering breaks
	}

	var b strings.Builder

	// === RESPONSIVE LOGO ===
	var logo string
	if contentWidth >= 60 {
		// Large logo for bigger terminals
		logo = `
 ███████╗ █████╗      ██╗     ██╗ █████╗ ██████╗ 
 ██╔════╝██╔══██╗     ██║     ██║██╔══██╗██╔══██╗
 ███████╗███████║     ██║     ██║███████║██║  ██║
 ╚════██║██╔══██║██   ██║██   ██║██╔══██║██║  ██║
 ███████║██║  ██║╚█████╔╝╚█████╔╝██║  ██║██████╔╝
 ╚══════╝╚═╝  ╚═╝ ╚════╝  ╚════╝ ╚═╝  ╚═╝╚═════╝ 
      █████╗ ██╗██╗   ██╗ ██████╗  ██████╗ ██████╗ 
     ██╔══██╗██║╚██╗ ██╔╝██╔═══██╗██╔═══██╗██╔══██╗
     ███████║██║ ╚████╔╝ ██║   ██║██║   ██║██████╔╝
     ██╔══██║██║  ╚██╔╝  ██║   ██║██║   ██║██╔══██╗
     ██║  ██║██║   ██║   ╚██████╔╝╚██████╔╝██████╔╝
     ╚═╝  ╚═╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝ ╚═════╝ 
`
	} else {
		// Smaller logo for narrow terminals
		logo = `
 ╔═╗┌─┐ ┬ ┬┌─┐┌┬┐
 ╚═╗├─┤ │ │├─┤ ││
 ╚═╝┴ ┴└┘└┘┴ ┴─┴┘
   ╔═╗┬┬ ┬┌─┐┌─┐┌┐ 
   ╠═╣│└┬┘│ ││ │├┴┐
   ╩ ╩┴ ┴ └─┘└─┘└─┘
`
	}
	b.WriteString(centerText(logoStyle.Render(logo), contentWidth))

	// === TAGLINE (properly centered) ===
	taglineText := "Backend Developer · Cloud Enthusiast · DevOps"
	taglineStyle := lipgloss.NewStyle().Foreground(muted).Italic(true)
	taglineRendered := taglineStyle.Render(taglineText)
	b.WriteString("\n")
	b.WriteString(centerText(taglineRendered, contentWidth))
	b.WriteString("\n\n")

	// === NAVIGATION ===
	var navItems []string
	if m.view == ViewList {
		navItems = append(navItems, navActive.Render("◆ home"))
	} else {
		navItems = append(navItems, navInactive.Render("◇ home"))
	}
	navItems = append(navItems, navInactive.Render("│"))
	if m.view == ViewDetail {
		navItems = append(navItems, navActive.Render("◆ details"))
	} else {
		navItems = append(navItems, navInactive.Render("◇ details"))
	}
	nav := lipgloss.JoinHorizontal(lipgloss.Center, navItems...)
	b.WriteString(centerText(nav, contentWidth))
	b.WriteString("\n")

	// === DIVIDER ===
	dividerWidth := contentWidth - 10
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	divider := lipgloss.NewStyle().Foreground(dimmed).Render(strings.Repeat("─", dividerWidth))
	b.WriteString(centerText(divider, contentWidth))
	b.WriteString("\n")

	if m.view == ViewList {
		// === PROJECT LIST ===
		projectSection := sectionStyle.Render("▸ PROJECTS")
		b.WriteString(centerText(projectSection, contentWidth))
		b.WriteString("\n\n")

		for i, item := range items {
			if item.Category == "projects" {
				icon := item.Icon
				tag := tagStyle.Render(item.Tag)

				if m.cursor == i {
					line := fmt.Sprintf("› %s %s  %s", icon, item.Title, tag)
					b.WriteString(centerText(itemSelected.Render(line), contentWidth))
				} else {
					line := fmt.Sprintf("  %s %s  %s", icon, item.Title, tag)
					b.WriteString(centerText(itemNormal.Render(line), contentWidth))
				}
				b.WriteString("\n")
			}
		}

		// === ABOUT SECTION ===
		b.WriteString("\n")
		aboutSection := sectionStyle.Render("▸ ABOUT")
		b.WriteString(centerText(aboutSection, contentWidth))
		b.WriteString("\n\n")

		for i, item := range items {
			if item.Category == "about" {
				icon := item.Icon
				tag := tagStyle.Render(item.Tag)

				if m.cursor == i {
					line := fmt.Sprintf("› %s %s  %s", icon, item.Title, tag)
					b.WriteString(centerText(itemSelected.Render(line), contentWidth))
				} else {
					line := fmt.Sprintf("  %s %s  %s", icon, item.Title, tag)
					b.WriteString(centerText(itemNormal.Render(line), contentWidth))
				}
				b.WriteString("\n")
			}
		}

		// === SOCIAL LINKS (Clickable!) ===
		b.WriteString("\n")
		socialSection := sectionStyle.Render("▸ CONNECT")
		b.WriteString(centerText(socialSection, contentWidth))
		b.WriteString("\n\n")

		for _, s := range socials {
			// Show URL directly (OSC8 hyperlinks don't work in all SSH clients)
			// Calculate padding manually to handle OSC8 sequences correctly
			visibleText := socialIcon.Render(s.Icon) + socialText.Render(s.URL)
			textWidth := lipgloss.Width(visibleText)
			padding := (contentWidth - textWidth) / 2
			if padding < 0 {
				padding = 0
			}

			socialLine := socialIcon.Render(s.Icon) + hyperlink(s.Link, socialText.Render(s.URL))
			b.WriteString(strings.Repeat(" ", padding) + socialLine)
			b.WriteString("\n")
		}

		// === QUOTE (Easter egg) ===
		if m.showQuote {
			b.WriteString("\n")
			quote := quoteStyle.Render(m.currentQuote)
			b.WriteString(centerText(quote, contentWidth))
			b.WriteString("\n")
		}

		// === CONFETTI (Easter egg) ===
		if m.showConfetti {
			confetti := []string{"🎉", "✨", "🎊", "⭐", "💫", "🌟"}
			var confettiLine string
			for i := 0; i < 10; i++ {
				confettiLine += confetti[rand.Intn(len(confetti))] + " "
			}
			b.WriteString("\n")
			b.WriteString(centerText(confettiLine, contentWidth))
		}

		// === HINTS ===
		b.WriteString("\n")
		hints := hintStyle.Render("↑↓ navigate · enter view · ? help · q quit")
		b.WriteString(centerText(hints, contentWidth))

	} else if m.view == ViewDetail {
		// === DETAIL VIEW ===
		item := items[m.cursor]

		// Adjust detail box width based on content width
		boxWidth := contentWidth - 4
		if boxWidth < 40 {
			boxWidth = 40
		}
		if boxWidth > 60 {
			boxWidth = 60
		}

		dynamicDetailBox := detailBox.Copy().Width(boxWidth)
		dynamicDescStyle := descStyle.Copy().Width(boxWidth - 6)

		// Project link
		var linkLine string
		if item.Link != "" {
			linkStyle := lipgloss.NewStyle().Foreground(cyan).Underline(true)
			visibleLink := linkStyle.Render("→ " + item.Link)
			linkLine = hyperlink(item.Link, visibleLink)
		}

		detailContent := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render(item.Icon+"  "+item.Title),
			"",
			techStyle.Render(item.TechStack),
			"",
			dynamicDescStyle.Render(item.Description),
			"",
			tagStyle.Render(" "+item.Tag+" "),
		)

		if linkLine != "" {
			detailContent = lipgloss.JoinVertical(lipgloss.Left, detailContent, "", linkLine)
		}

		box := dynamicDetailBox.Render(detailContent)
		b.WriteString("\n")
		b.WriteString(centerText(box, contentWidth))
		b.WriteString("\n\n")

		hints := hintStyle.Render("esc back · q quit")
		b.WriteString(centerText(hints, contentWidth))
	} else if m.view == ViewHelp {
		// === HELP / SECRETS ===
		helpSection := sectionStyle.Render("▸ COMMANDS & SECRETS")
		b.WriteString(centerText(helpSection, contentWidth))
		b.WriteString("\n\n")

		secrets := []struct{ key, desc string }{
			{"↑ / k", "Navigate up"},
			{"↓ / j", "Navigate down"},
			{"enter / spc", "Open details"},
			{"esc / bksp", "Go back"},
			{"q", "Quit"},
			{"?", "Toggle help"},
			{"s", "A little surprise"},
			{"type 'hello'", "Say hello"},
			{"type 'hire'", "Hiring info"},
		}

		for _, s := range secrets {
			keyStr := tagStyle.Render(s.key)
			descStr := itemNormal.Render(s.desc)
			fullLine := keyStr + "   " + descStr
			b.WriteString(centerText(fullLine, contentWidth))
			b.WriteString("\n\n")
		}

		hints := hintStyle.Render("esc back · q quit")
		b.WriteString(centerText(hints, contentWidth))
	}

	// === FOOTER ===
	b.WriteString("\n")
	// OLD STATIC FOOTER:
	// footerText := lipgloss.NewStyle().Foreground(dimmed).Render("━━━ © 2026 ━━━")

	// NEW DYNAMIC FOOTER (Uses your hints!):
	// Pick a hint based on the seconds of the current time so it rotates
	hintIndex := int(time.Now().Unix() % int64(len(easterEggHints)))
	hint := easterEggHints[hintIndex]
	footerText := lipgloss.NewStyle().Foreground(dimmed).Render(fmt.Sprintf("━━━ © 2026 ━━━ %s ━━━", hint))
	b.WriteString(centerText(footerText, contentWidth))

	// === MAIN BORDER BOX ===
	mainBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(accent).
		Padding(1, 1). // Reduced padding
		Width(contentWidth).
		Render(b.String())

	// Center the box in the terminal
	boxHeight := lipgloss.Height(mainBox)
	verticalPadding := 0
	if height > boxHeight {
		verticalPadding = (height - boxHeight) / 2
	}

	horizontalPadding := 0
	boxWidth := lipgloss.Width(mainBox)
	if width > boxWidth {
		horizontalPadding = (width - boxWidth) / 2
	}

	return lipgloss.NewStyle().
		MarginTop(verticalPadding).
		MarginLeft(horizontalPadding).
		Render(mainBox)
}
func init() {
	// Force "True Color" (24-bit) output, bypassing environment checks
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// ... func main() starts here

// --- 6. SERVER ---

func main() {
	rand.Seed(time.Now().UnixNano())
	s, err := wish.NewServer(
		wish.WithAddress("0.0.0.0:23234"),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			wb.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return initialModel(), []tea.ProgramOption{tea.WithAltScreen()}
			}),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on port 23234...")

	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
