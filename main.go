package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

var systemStyle = lg.NewStyle().
	BorderStyle(lg.DoubleBorder()).BorderForeground(lg.Color("#33ffaa")).
	Background(lg.Color("#000087")).
	Foreground(lg.Color("#ffffff")).
	Padding(2, 4)

type models struct {
	UIControl   UIControl
	Notes       []*Note
	SectionData []Section
}

func initialModel() models {
	mockNotes := []*Note{} // Create a slice with capacity for 4 items
	for i := range 4 {
		mockNotes = append(mockNotes, NewMockNote("test"+strconv.Itoa(i), i))
	}

	return models{
		Notes: mockNotes,
		SectionData: []Section{
			{ID: 0, Order: 0, Name: "Uncategorized"},
		},
	}
}

func (m models) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."

	/*
		Maybe check if there is section that I otherwise create the uncategorized one.
	*/
	return nil
}

func (m models) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	/* TODO
	Update display order
	right now reorder is not supported. it is sort by time
	*/

	/*
		Group Notes with the same ID to Display order
	*/
	m.UIControl.DisplayOrder = make(map[int][]*Note)
	dp := m.UIControl.DisplayOrder
	for _, notePtr := range m.Notes {
		dp[notePtr.SectionID] = append(dp[notePtr.SectionID], notePtr)
	}

	/*
		Lets think about the algo
		concern:
			0. There will be time where all order will be zero  <- this likely won't happen when you connect to database
					so lets just skip this logic and just mock them
					TODO: mock order data when init

			1. element might get add later with zero value order
					This also mean I should keep current Max value globally at least in that section

	*/

	// *** DO THIS when ADD and REORDER
	// for sectionID, notes := range dp {
	// 	currMax := slices.MaxFunc(notes, func(a, b *Note) int {
	// 		return a.Order - b.Order
	// 	}).Order

	// 	sort.Slice(notes, func(i, j int) bool {
	// 		return notes[i].DateCreated
	// 	})
	// }

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UIControl.TermSize.Height = msg.Height
		m.UIControl.TermSize.Width = msg.Width
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.UIControl.RowCursor > 0 {
				m.UIControl.RowCursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			//TODO Scope len() to each section column << DONE?
			if notes := dp[m.UIControl.SectionCursor]; m.UIControl.RowCursor < len(notes)-1 {
				m.UIControl.RowCursor++
			}

		case "left", "h":
			if m.UIControl.SectionCursor > 0 {
				m.UIControl.SectionCursor--
				m.UIControl.RowCursor = 0
			}

		// The "down" and "j" keys move the cursor down
		case "right", "l":
			//TODO Scope len() to each section column << DONE?
			if m.UIControl.SectionCursor < len(dp)-1 {
				m.UIControl.SectionCursor++
				m.UIControl.RowCursor = 0
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			sectionNotePtrs, ok := m.UIControl.DisplayOrder[m.UIControl.SectionCursor]
			if ok {
				notePtr := sectionNotePtrs[m.UIControl.RowCursor]
				notePtr.IsChecked = !notePtr.IsChecked
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m models) View() string {

	// The header
	s := "What should we buy at the market?\n\n"

	dpo := m.UIControl.DisplayOrder
	sectionIDs := maps.Keys(dpo)

	sectionList := []Section{}
	for sid := range sectionIDs {

		idx := slices.IndexFunc(m.SectionData, func(el Section) bool {
			return el.ID == sid
		})

		if idx > -1 {
			section := m.SectionData[idx]
			sectionList = append(sectionList, section)
		}
	}
	//now we have section data

	// sort section id by SectionData.Order
	sort.Slice(sectionList, func(i, j int) bool {
		return sectionList[i].Order < sectionList[j].Order
	})

	slices.SortFunc(sectionList, func(a, b Section) int {
		return a.Order - b.Order
	})

	// Iterate over our choices
	for i, section := range sectionList {
		notesInSection := dpo[section.ID]

		if len(notesInSection) == 0 {
			continue
		}

		sortedNotes := slices.Clone(notesInSection)
		slices.SortFunc(sortedNotes, func(a, b *Note) int {
			return a.Order - b.Order
		})

		for _, note := range sortedNotes {
			// Is the cursor pointing at this item?
			cursor := " " // no cursor
			if m.UIControl.RowCursor == note.Order && m.UIControl.SectionCursor == section.Order {
				cursor = ">" // cursor!
			}

			// Is this item selected?
			checked := " " // not selected
			if note.IsChecked {
				checked = "x"
			}

			tmpS := fmt.Sprintf("%s [%s] %s", cursor, checked, note.Content)

			// Render the row
			if m.UIControl.RowCursor == i {
				tmpS = lg.NewStyle().Foreground(lg.Color("23")).Render(tmpS)
			}

			s += tmpS
			s += "\n"
		}
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return systemStyle.Width(m.UIControl.TermSize.Width - 3).Height(m.UIControl.TermSize.Height - 5).Render(s)

}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// Notes
type Note struct {
	ID          int       // Database ID, unique for each Note
	Order       int       // Display order of the note
	Content     string    // The content of the note
	SectionID   int       // Pointer to the parent Section
	DateUpdated time.Time // Timestamp when the note was last updated
	DateCreated time.Time // Timestamp when the note was created
	IsChecked   bool      // Is the note completed/checked?
	IsDeleted   bool      // Flag for soft deletion
}

func NewNote(content string) *Note {
	return &Note{
		Content:     content,
		DateUpdated: time.Now(),
		DateCreated: time.Now(),
	}
}

func NewMockNote(content string, order int) *Note {
	return &Note{
		Content:     content,
		DateUpdated: time.Now(),
		DateCreated: time.Now(),
		Order:       order,
	}
}

type Section struct {
	ID    int    // Unique identifier for the Section
	Order int    // Display order
	Name  string // Section name
}

type UIControl struct {
	IsDialogOpened bool            // Tracks if a dialog is open
	LastUIBuffer   string          // Stores the last state or buffer for UI
	DisplayOrder   map[int][]*Note // Map SectionIDs to corresponding Notes
	RowCursor      int             // which to-do list item our cursor is pointing at in a section
	SectionCursor  int             // which column(Section) our cursor is pointing at
	TermSize       struct {        // terminal size. currently use for fullscreen
		Width  int
		Height int
	}
}
