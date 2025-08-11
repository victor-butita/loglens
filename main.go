package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Step 1: Define the Model ---
type model struct {
	logEntries    []map[string]interface{}
	selectedIndex int // Index of the log entry we're currently looking at
	width         int // Width of the terminal
	height        int // Height of the terminal
}

// Init is called once when the program starts.
func (m model) Init() tea.Cmd {
	return watchLogFile()
}

// --- Step 2: Define Messages ---
type allLogsReadMsg struct {
	entries []map[string]interface{}
}
// This message is sent when the terminal window is resized.
type windowSizeMsg struct {
	Width  int
	Height int
}

// This command reads the entire file and returns one single message with all the data.
func watchLogFile() tea.Cmd {
	// For this step, we'll read the file synchronously.
	// In a future step, we could make this a real-time goroutine.
	return func() tea.Msg {
		file, err := os.Open("logs.jsonl")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
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
	switch msg := msg.(type) {
	// A message that the window has been resized.
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		// Scroll up
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		// Scroll down
		case "down", "j":
			if m.selectedIndex < len(m.logEntries)-1 {
				m.selectedIndex++
			}
		}

	case allLogsReadMsg:
		m.logEntries = msg.entries
	}

	return m, nil
}

// --- Step 4: Define the View function ---
// This is where all the new UI logic goes.
func (m model) View() string {
	if len(m.logEntries) == 0 {
		return "Reading logs..."
	}

	// --- Define Styles using Lipgloss ---
	// We'll have a style for the selected list item and the detail view pane.
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("212")).Foreground(lipgloss.Color("0"))
	detailStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	
	// --- Build the Left Pane (List of Logs) ---
	var listBuilder strings.Builder
	for i, log := range m.logEntries {
		// Try to get level and message for a nice summary
		level, _ := log["level"].(string)
		message, _ := log["message"].(string)
		summary := fmt.Sprintf("[%s] %s", strings.ToUpper(level), message)
		
		if i == m.selectedIndex {
			// Apply selected style
			listBuilder.WriteString(selectedStyle.Render(summary))
		} else {
			// Normal item
			listBuilder.WriteString(summary)
		}
		listBuilder.WriteString("\n")
	}

	// --- Build the Right Pane (Detail View) ---
	var detailBuilder strings.Builder
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.logEntries) {
		selectedLog := m.logEntries[m.selectedIndex]
		prettyJSON, err := json.MarshalIndent(selectedLog, "", "  ")
		if err != nil {
			detailBuilder.WriteString("Error rendering JSON")
		} else {
			detailBuilder.WriteString(string(prettyJSON))
		}
	}

	// --- Combine Panes using Lipgloss ---
	// Calculate pane widths. Let's give 1/3 to the list and 2/3 to the detail.
	listWidth := m.width / 3
	detailWidth := m.width - listWidth

	// Set the height of our panes.
	// We subtract a few lines for potential headers/footers later.
	paneHeight := m.height - 4 
	
	listPane := lipgloss.NewStyle().
		Width(listWidth).
		Height(paneHeight).
		Render(listBuilder.String())

	detailPane := detailStyle.
		Width(detailWidth).
		Height(paneHeight).
		Render(detailBuilder.String())

	// Join the two panes horizontally.
	return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
}

// --- Step 5: Run the application ---
func main() {
	p := tea.NewProgram(
		model{logEntries: []map[string]interface{}{}, selectedIndex: 0},
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}