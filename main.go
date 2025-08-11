package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// --- Step 1: Define the Model ---
type model struct {
	logEntries      []map[string]interface{}
	filteredEntries []map[string]interface{} // The logs that match the filter
	selectedIndex   int
	filterInput     textinput.Model // The text input component for our filter
	glamourRenderer *glamour.TermRenderer // For syntax highlighting JSON
	width           int
	height          int
}

// Helper function to create the initial model
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Filter logs..."
	ti.Focus() // Make the text input active
	ti.CharLimit = 156
	ti.Width = 50

	// Setup the glamour renderer
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0), // We handle word wrap ourselves
	)

	return model{
		logEntries:      []map[string]interface{}{},
		filteredEntries: []map[string]interface{}{},
		selectedIndex:   0,
		filterInput:     ti,
		glamourRenderer: renderer,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, watchLogFile())
}

// --- Step 2: Define Messages ---
type allLogsReadMsg struct{ entries []map[string]interface{} }

func watchLogFile() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open("logs.jsonl")
		if err != nil { return nil }
		defer file.Close()

		var allEntries []map[string]interface{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
				allEntries = append(allEntries, logEntry)
			}
		}
		return allLogsReadMsg{entries: allEntries}
	}
}

// --- Step 3: Define the Update function ---
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			// Use filteredEntries for boundary check
			if m.selectedIndex < len(m.filteredEntries)-1 {
				m.selectedIndex++
			}
		}

	case allLogsReadMsg:
		m.logEntries = msg.entries
		m.filteredEntries = msg.entries // Initially, all logs are shown
	}

	// Handle text input updates
	m.filterInput, cmd = m.filterInput.Update(msg)

	// After any input, re-filter the logs
	filter := m.filterInput.Value()
	var newFiltered []map[string]interface{}
	if filter == "" {
		newFiltered = m.logEntries
	} else {
		for _, log := range m.logEntries {
			logStr, _ := json.Marshal(log)
			if strings.Contains(strings.ToLower(string(logStr)), strings.ToLower(filter)) {
				newFiltered = append(newFiltered, log)
			}
		}
	}
	m.filteredEntries = newFiltered
	// Reset selectedIndex if it's out of bounds after filtering
	if m.selectedIndex >= len(m.filteredEntries) && len(m.filteredEntries) > 0 {
		m.selectedIndex = len(m.filteredEntries) - 1
	} else if len(m.filteredEntries) == 0 {
		m.selectedIndex = 0
	}


	return m, cmd
}

// --- Step 4: Define the View function ---
func (m model) View() string {
	if len(m.logEntries) == 0 {
		return "Reading logs..."
	}

	// --- Header (Filter Input) ---
	header := "LogLens - " + m.filterInput.View() + "\n"

	// --- Left Pane (Filtered List) ---
	var listBuilder strings.Builder
	for i, log := range m.filteredEntries {
		level, _ := log["level"].(string)
		message, _ := log["message"].(string)
		summary := fmt.Sprintf("[%s] %s", strings.ToUpper(level), message)
		
		// Truncate summary if it's too long
		listWidth := m.width / 3
		if len(summary) > listWidth {
			summary = summary[:listWidth-3] + "..."
		}

		if i == m.selectedIndex {
			listBuilder.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("212")).Render(summary))
		} else {
			listBuilder.WriteString(summary)
		}
		listBuilder.WriteString("\n")
	}

	// --- Right Pane (Detail View with Color) ---
	var detailStr string
	if len(m.filteredEntries) > 0 && m.selectedIndex < len(m.filteredEntries) {
		selectedLog := m.filteredEntries[m.selectedIndex]
		prettyJSON, _ := json.MarshalIndent(selectedLog, "", "  ")
		
		// Use Glamour to render the JSON with syntax highlighting
		// We tell it it's a JSON code block.
		jsonForGlamour := "```json\n" + string(prettyJSON) + "\n```"
		detailStr, _ = m.glamourRenderer.Render(jsonForGlamour)
	} else {
		detailStr = "No logs match the filter."
	}
	
	// --- Combine ---
	listPane := lipgloss.NewStyle().Width(m.width / 3).Render(listBuilder.String())
	detailPane := lipgloss.NewStyle().Width(m.width - m.width/3).Render(detailStr)
	
	// Join the header and the panes vertically.
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
	return lipgloss.JoinVertical(lipgloss.Left, header, mainContent)
}


// --- Step 5: Run the application ---
func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}