package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

const (
	CardBorderColor = "#6495ED"
)

var mockId = 10

var systemStyle = lg.NewStyle().
	BorderStyle(lg.DoubleBorder()).BorderForeground(lg.Color("#33ffaa")).
	// Background(lg.Color("#71797E")).
	// Foreground(lg.Color("#ffffff"))
	Padding(1, 2)

var cardStyle = lg.NewStyle().
	BorderStyle(lg.RoundedBorder()).BorderForeground(lg.Color(CardBorderColor)).
	// BorderBackground(lg.Color("#FFBF00")).
	// Background(lg.Color("#FFBF00")).
	// Foreground(lg.Color("#ffffff")).
	Padding(1, 2, 1, 0).
	Height(4).Width(20)

var selectedStyle = lg.NewStyle().Inherit(cardStyle).
	BorderStyle(lg.DoubleBorder()).BorderForeground(lg.Color(CardBorderColor)).
	Padding(1, 2, 1, 0)

var bold = lg.NewStyle().Bold(true)

// var bgStyle = lg.NewStyle().Background(lg.Color("#FFBF00"))

type models struct {
	UIControl   UIControl
	Notes       []*Note
	SectionData []Section
	isInit      bool
}

func initialModel() models {
	mockNotes := []*Note{} // Create a slice with capacity for 4 items
	for i := range 4 {
		mockNotes = append(mockNotes, NewMockNote("test"+strconv.Itoa(i), i, 0))
	}
	for i := range 2 {
		mockNotes = append(mockNotes, NewMockNote("test"+strconv.Itoa(i), i, 1))
	}

	return models{
		Notes: mockNotes,
		SectionData: []Section{
			{ID: 0, Order: 0, Name: "Uncategorized"},
			{ID: 1, Order: 1, Name: "Inbox"},
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
	/*
		Group Notes with the same ID to Display order
	*/

	if !m.isInit {
		m.RepopulateDisplayOrder()
		m.isInit = true
	}
	dp := m.UIControl.DisplayOrder
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
			section, ok := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
			if !ok {
				return m, nil
			}
			if notes := dp[section.ID]; m.UIControl.RowCursor < len(notes)-1 {
				m.UIControl.RowCursor++
			}

		// The "left" and "h" keys move the cursor left to the previous section
		case "left", "h":
			if m.UIControl.SectionCursor > 0 {
				m.UIControl.SectionCursor--

				section, ok := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
				if !ok {
					return m, nil
				}
				notesCount := len(dp[section.ID])
				if m.UIControl.RowCursor > notesCount-1 {
					m.UIControl.RowCursor = notesCount - 1
				}
				if m.UIControl.RowCursor < 0 {
					m.UIControl.RowCursor = 0
				}
			}

		// The "left" and "h" keys move the cursor right to the next section
		case "right", "l":
			if m.UIControl.SectionCursor < len(dp)-1 {
				m.UIControl.SectionCursor++

				section, ok := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
				if !ok {
					return m, nil
				}
				if notesCount := len(dp[section.ID]); m.UIControl.RowCursor > notesCount-1 {
					m.UIControl.RowCursor = notesCount - 1
				}
				if m.UIControl.RowCursor < 0 {
					m.UIControl.RowCursor = 0
				}
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			section, ok := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
			if !ok {
				return m, nil
			}

			sectionNotePtrs, ok := m.UIControl.DisplayOrder[section.ID]
			if ok {
				if len(sectionNotePtrs) > 0 {
					notePtr := sectionNotePtrs[m.UIControl.RowCursor]
					notePtr.IsChecked = !notePtr.IsChecked
				}
			}

		case "a":
			section, ok := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
			if !ok {
				return m, nil
			}
			sectionNotePtrs, ok := m.UIControl.DisplayOrder[section.ID]
			if ok {
				maxOrder := -1
				if len(sectionNotePtrs) > 0 {
					maxOrder = slices.MaxFunc(sectionNotePtrs, func(a, b *Note) int {
						return a.Order - b.Order
					}).Order
				} else {
					//Hijack this control flow to fix cursor after add notes in an empty section
					m.UIControl.RowCursor = 0
				}

				sectionIdx := slices.IndexFunc(m.SectionData, func(sec Section) bool {
					return sec.Order == m.UIControl.SectionCursor
				})

				/*TODO There must be better way to link the data. This manually delete is not bad but
				I have a feeling that it can be better
				*/
				/*
					Maybe I should just manipulate the original array and repopulate displayOrder every Update
				*/
				if sectionIdx != -1 {
					sec := m.SectionData[sectionIdx]
					buffer := NewMockNote("test"+strconv.Itoa(maxOrder+1), maxOrder+1, sec.ID)
					tmp := append(sectionNotePtrs, buffer)
					m.UIControl.DisplayOrder[section.ID] = tmp

					m.Notes = slices.DeleteFunc(m.Notes, func(n *Note) bool {
						return n.SectionID == sec.ID
					})

					m.Notes = append(m.Notes, tmp...)
				}
			}

		case "d":
			section, ok := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
			if !ok {
				return m, nil
			}
			if _, ok = m.UIControl.DisplayOrder[section.ID]; ok {
				sectionIdx := slices.IndexFunc(m.SectionData, func(sec Section) bool {
					return sec.Order == m.UIControl.SectionCursor
				})

				if sectionIdx != -1 {
					sec := m.SectionData[sectionIdx]

					m.Notes = slices.DeleteFunc(m.Notes, func(n *Note) bool {
						return n.Order == m.UIControl.RowCursor && n.SectionID == sec.ID
					})

					if m.UIControl.RowCursor > 0 {
						m.UIControl.RowCursor--
					}
					m.RepopulateDisplayOrder()
					RecalulateNoteOrder(m.UIControl.DisplayOrder[sec.ID])
				}
			}

		case "A":
			maxOrder := -1
			if len(m.SectionData) > 0 {
				maxOrder = slices.MaxFunc(m.SectionData, func(a, b Section) int {
					return a.Order - b.Order
				}).Order
			}
			mockId++
			m.SectionData = append(m.SectionData, NewMockSection("TestSection", maxOrder+1, mockId))
			m.RepopulateDisplayOrder()

		case "D":

			if len(m.SectionData) == 1 {
				break
			}

			sectionIdx := slices.IndexFunc(m.SectionData, func(sec Section) bool {
				return sec.Order == m.UIControl.SectionCursor
			})

			if sectionIdx != -1 {
				sec := m.SectionData[sectionIdx]
				//Delete Notes Data
				m.Notes = slices.DeleteFunc(m.Notes, func(n *Note) bool {
					return n.SectionID == sec.ID
				})
			}

			//Delete Section Data
			// m.SectionData = slices.DeleteFunc(m.SectionData, func(sec Section) bool { return sec.Order == m.UIControl.SectionCursor })
			m.SectionData = slices.Delete(m.SectionData, sectionIdx, sectionIdx+1)
			//Recalculate Section Order
			m.RepopulateDisplayOrder()
			RecalulateSectionOrder(m.SectionData)

			if m.UIControl.SectionCursor > 0 {
				m.UIControl.SectionCursor--
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func FindSectionByOrder(data []Section, order int) (Section, bool) {
	sIX := slices.IndexFunc(data, func(s Section) bool {
		return s.Order == order
	})

	if sIX != -1 {
		return data[sIX], true
	} else {
		return Section{}, false
	}
}

func RecalulateNoteOrder(notes []*Note) {
	slices.SortFunc(notes, func(a, b *Note) int {
		return a.Order - b.Order
	})

	for i := range len(notes) {
		notes[i].Order = i
	}

}

func RecalulateSectionOrder(section []Section) {
	slices.SortFunc(section, func(a, b Section) int {
		return a.Order - b.Order
	})

	for i := range len(section) {
		section[i].Order = i
	}

}

func (m *models) RepopulateDisplayOrder() {
	m.UIControl.DisplayOrder = make(map[int][]*Note)
	dp := m.UIControl.DisplayOrder

	for _, section := range m.SectionData {
		dp[section.ID] = make([]*Note, 0)
	}
	for _, notePtr := range m.Notes {
		dp[notePtr.SectionID] = append(dp[notePtr.SectionID], notePtr)
	}
}

func (m models) View() string {

	// The header
	allText := ""

	dpo := m.UIControl.DisplayOrder
	sectionIDs := maps.Keys(dpo)

	// return spew.Sdump(dpo)

	//Collect Section Data
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
	slices.SortFunc(sectionList, func(a, b Section) int {
		return a.Order - b.Order
	})

	// Iterate over our sections
	for loopCnt, section := range sectionList {
		notesInSection := dpo[section.ID]

		sectionText := ""

		if m.UIControl.SectionCursor == section.Order {
			sectionText = bold.Underline(true).Render(section.Name + strconv.Itoa(section.ID))
		} else {
			sectionText = bold.Render(section.Name + strconv.Itoa(section.ID))
		}

		sectionText += "\n\n"
		sortedNotes := slices.Clone(notesInSection)
		slices.SortFunc(sortedNotes, func(a, b *Note) int {
			return a.Order - b.Order
		})

		// Iterate over our sortedNotes in the section
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

			tmpS := fmt.Sprintf(" %s[%s] %s", cursor, checked, note.Content)

			if m.UIControl.RowCursor == note.Order && m.UIControl.SectionCursor == section.Order {
				tmpS = selectedStyle.Render(tmpS)
			} else {
				tmpS = cardStyle.Render(tmpS)
			}

			// s = lg.JoinVertical(lg.Bottom, s, tmpS)
			sectionText += tmpS
			sectionText += "\n"
		}

		allText = lg.JoinHorizontal(lg.Top, allText, sectionText)
		if loopCnt < len(sectionList)-1 {
			allText = lg.JoinHorizontal(lg.Top, allText, "        ")
		}
	}

	// The footer
	allText += "\nPress q to quit.\n"

	// DEBUG
	// allText += spew.Sdump(m.SectionData)
	// allText += fmt.Sprintf("RowCursor = %d,SectionCursor= %d\n", m.UIControl.RowCursor, m.UIControl.SectionCursor)
	// allText += fmt.Sprintf("DisplayOrder Len = %d\n", len(m.UIControl.DisplayOrder))
	// for _, s := range m.SectionData {
	// 	allText += strconv.Itoa(s.ID)
	// 	allText += ", "
	// }
	// section, _ := FindSectionByOrder(m.SectionData, m.UIControl.SectionCursor)
	// allText += spew.Sdump(m.UIControl.DisplayOrder[section.ID])

	// Send the UI for rendering
	return systemStyle.Width(m.UIControl.TermSize.Width - 3).Height(m.UIControl.TermSize.Height - 5).Render(allText)

}

func mapSlice(input []int, transform func(int) int) []int {
	result := make([]int, len(input))
	for i, v := range input {
		result[i] = transform(v)
	}
	return result
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return "#" + hex.EncodeToString(bytes)
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

func NewMockNote(content string, order int, sectionId int) *Note {
	return &Note{
		Content:     content,
		DateUpdated: time.Now(),
		DateCreated: time.Now(),
		Order:       order,
		SectionID:   sectionId,
	}
}

func NewMockSection(content string, order int, sectionId int) Section {
	return Section{
		Name:  content,
		ID:    sectionId,
		Order: order,
	}
}

type Section struct {
	ID    int    // Unique identifier for the Section
	Order int    // Display order
	Name  string // Section name
	// HighestNoteOrder int    // Store height notes order for ordering
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
