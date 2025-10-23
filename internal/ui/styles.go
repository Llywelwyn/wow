package ui

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

type palette struct {
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Muted     lipgloss.AdaptiveColor
	Success   lipgloss.AdaptiveColor
	Danger    lipgloss.AdaptiveColor
	Text      lipgloss.AdaptiveColor
}

type colors struct {
	Primary   string
	Secondary string
	Muted     string
	Success   string
	Danger    string
	Text      string
}

func defaultPalette() palette {
	light := colors{
		Primary:   "#0066CC",
		Secondary: "#6B46C1",
		Muted:     "#64748B",
		Success:   "#059669",
		Danger:    "#DC2626",
		Text:      "#1E293B",
	}

	dark := colors{
		Primary:   "#66B3FF",
		Secondary: "#A78BFA",
		Muted:     "#94A3B8",
		Success:   "#34D399",
		Danger:    "#F87171",
		Text:      "#F1F5F9",
	}

	return palette{
		Primary:   lipgloss.AdaptiveColor{Light: light.Primary, Dark: dark.Primary},
		Secondary: lipgloss.AdaptiveColor{Light: light.Secondary, Dark: dark.Secondary},
		Muted:     lipgloss.AdaptiveColor{Light: light.Muted, Dark: dark.Muted},
		Success:   lipgloss.AdaptiveColor{Light: light.Success, Dark: dark.Success},
		Danger:    lipgloss.AdaptiveColor{Light: light.Danger, Dark: dark.Danger},
		Text:      lipgloss.AdaptiveColor{Light: light.Text, Dark: dark.Text},
	}
}

type Styles struct {
	// Textual
	Heading   lipgloss.Style
	Body      lipgloss.Style
	Muted     lipgloss.Style
	Bold      lipgloss.Style
	Underline lipgloss.Style
	Italic    lipgloss.Style

	// Semantic
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style

	// Components
	Label lipgloss.Style
	Key   lipgloss.Style
	Value lipgloss.Style
	Tag   lipgloss.Style
	Icon  lipgloss.Style

	// State
	Empty    lipgloss.Style
	Disabled lipgloss.Style
}

var (
	styles Styles
	once   sync.Once
)

func GetStyles() Styles {
	once.Do(initDefaultStyle)
	return styles
}

func initDefaultStyle() {
	p := defaultPalette()
	initStyle(p)
}

func initStyle(p palette) {
	base := lipgloss.NewStyle()

	styles = Styles{
		Heading:   base.Foreground(p.Primary).Bold(true),
		Body:      base.Foreground(p.Text),
		Muted:     base.Foreground(p.Muted),
		Bold:      base.Foreground(p.Text).Bold(true),
		Underline: base.Foreground(p.Text).Underline(true),
		Italic:    base.Foreground(p.Text).Italic(true),

		Primary:   base.Foreground(p.Primary),
		Secondary: base.Foreground(p.Secondary),
		Success:   base.Foreground(p.Success),
		Error:     base.Foreground(p.Danger),

		Label: base.Foreground(p.Muted).Bold(true),
		Key:   base.Foreground(p.Text).Bold(true),
		Value: base.Foreground(p.Primary),
		Tag:   base.Foreground(p.Secondary).Underline(true),
		Icon:  base.Foreground(p.Primary),

		Empty:    base.Foreground(p.Muted).Italic(true),
		Disabled: base.Foreground(p.Muted).Faint(true),
	}
}
