package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

var mockId = 10

func NewTextInputSetting() textinput.Model {
	ti := textinput.New()
	ti.CharLimit = 40
	ti.Width = 40

	return ti
}

func (m ProgramModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."

	/*
		Maybe check if there is section that I otherwise create the uncategorized one.
	*/
	return textinput.Blink
}

func (m ProgramModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	/*
		Group Notes with the same ID to Display order
	*/
	var cmd tea.Cmd

	if !m.IsInit {
		m.RepopulateDisplayOrder()
		m.IsInit = true
	}
	m.StatusText = ""
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

	if m.IsTextInputShown {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.UIControl.TermSize.Height = msg.Height
			m.UIControl.TermSize.Width = msg.Width

		// Is it a key press?
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				switch m.Operation {
				case "ADDNOTE":
					{
						m.TextInput.Blur()
						content := m.TextInput.Value()
						AddNote(&m, content)
						m.TextInput.SetValue("")
						if notes, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor); ok {
							m.UIControl.RowCursor = len(notes) - 1
						}
					}
				case "EDITNOTE":
					{
						m.TextInput.Blur()
						content := m.TextInput.Value()
						note := FindNoteByBothOrder(m, m.UIControl.SectionCursor, m.UIControl.RowCursor)
						if note != nil {
							EditNote(note, content)
							note.DateUpdated = time.Now()
						}
						m.TextInput.SetValue("")
					}
				case "ADDSECTION":
					{
						m.TextInput.Blur()
						name := m.TextInput.Value()
						name = strings.TrimSpace(name)
						if name == "" {
							name = "Unnamed Section"
						}

						maxOrder := -1
						if len(m.SectionData) > 0 {
							maxOrder = slices.MaxFunc(m.SectionData, func(a, b Section) int {
								return a.Order - b.Order
							}).Order
						}
						mockId++
						m.SectionData = append(m.SectionData, NewSection(name, maxOrder+1, mockId))
						m.UIControl.SectionCursor = maxOrder + 1
						m.RepopulateDisplayOrder()
					}

				case "EDITSECTION":
					{
						m.TextInput.Blur()
						name := m.TextInput.Value()
						section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
						if ok {
							EditSection(section, name)
						}

						m.TextInput.SetValue("")
						m.RepopulateDisplayOrder()
					}
				}

				//Reset to default. ready for new Operation
				m.Operation = ""
				m.IsTextInputShown = false

			case "esc":
				m.Operation = ""
				m.IsTextInputShown = false
				m.TextInput.SetValue("")
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd

	} else {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.UIControl.TermSize.Height = msg.Height
			m.UIControl.TermSize.Width = msg.Width

		// Is it a key press?
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit

			// The "up" and "k" keys move the cursor up
			case "up", "k":
				if m.UIControl.RowCursor > 0 {
					m.UIControl.RowCursor--
				}

			// The "down" and "j" keys move the cursor down
			case "down", "j":
				section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
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

					section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
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

					section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
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
				section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
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
				m.Operation = "ADDNOTE"
				m.IsTextInputShown = true
				m.InputPrompt = "What is the content of the note?"
				m.TextInput.Placeholder = "Type note content here"
				m.TextInput.SetValue("")
				m.TextInput, cmd = m.TextInput.Update(nil)
				m.TextInput.Focus()
				return m, cmd

			case "e":
				m.Operation = "EDITNOTE"
				m.IsTextInputShown = true
				m.InputPrompt = "What is the content of the note?"
				m.TextInput.Placeholder = "Type note content here"
				note := FindNoteByBothOrder(m, m.UIControl.SectionCursor, m.UIControl.RowCursor)
				if note != nil {
					m.TextInput.SetValue(note.Content)
				} else {
					m.TextInput.SetValue("")
				}
				m.TextInput, cmd = m.TextInput.Update(nil)
				m.TextInput.Focus()
				return m, cmd

			case "d":
				section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
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
				m.Operation = "ADDSECTION"
				m.IsTextInputShown = true
				m.InputPrompt = "What is the name of this section?"
				m.TextInput.Placeholder = "Type the section's name here"
				m.TextInput.SetValue("")
				m.TextInput, cmd = m.TextInput.Update(nil)
				m.TextInput.Focus()
				return m, cmd

			case "E":
				m.Operation = "EDITSECTION"
				m.IsTextInputShown = true
				m.InputPrompt = "What is the name of this section?"
				m.TextInput.Placeholder = "Type the section's name here"

				if section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor); ok {
					m.TextInput.SetValue(section.Name)
				} else {
					m.TextInput.SetValue("")
				}
				m.TextInput, cmd = m.TextInput.Update(nil)
				m.TextInput.Focus()
				return m, cmd

			case "D":
				{
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
			case "ctrl+s":
				{
					saveData := map[string]interface{}{
						"SectionData": m.SectionData,
						"Notes":       m.Notes,
					}

					jsonData, err := json.MarshalIndent(saveData, "", "    ")
					if err != nil {
						m.Debug = err.Error()
					}

					err = os.WriteFile("./data/save_file.json", jsonData, 0600)

					if err != nil {
						m.Debug = err.Error()
					}

					m.StatusText = "Data Saved!"
				}
			case "ctrl+r":
				{
					m = LoadMockData()
					m.TextInput = NewTextInputSetting()
				}
			}

		}
	}

	m.RepopulateDisplayOrder()
	return m, cmd
}

func AddNote(m *ProgramModel, content string) bool {
	section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor)
	if !ok {
		return false
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
			buffer := NewNote(content, maxOrder+1, sec.ID)
			tmp := append(sectionNotePtrs, buffer)
			m.UIControl.DisplayOrder[section.ID] = tmp

			m.Notes = slices.DeleteFunc(m.Notes, func(n *Note) bool {
				return n.SectionID == sec.ID
			})

			m.Notes = append(m.Notes, tmp...)
		}
	}
	return true
}

func EditNote(note *Note, content string) bool {
	if note != nil {
		note.Content = content
		return true
	}
	return false
}

func EditSection(section *Section, content string) bool {
	if section != nil {
		section.Name = content
		return true
	}
	return false
}

func FindNoteByBothOrder(m ProgramModel, sectionOrder int, noteOrder int) *Note {
	section, ok := FindSectionDataByOrder(m.SectionData, sectionOrder)
	if !ok {
		return nil
	}

	if sectionNotePtrs, ok := m.UIControl.DisplayOrder[section.ID]; ok {
		if note := FindNoteByItsOrder(sectionNotePtrs, m.UIControl.RowCursor); note != nil {
			return note
		}
	}

	return nil
}

func FindSectionDataByOrder(data []Section, order int) (*Section, bool) {
	sIX := slices.IndexFunc(data, func(s Section) bool {
		return s.Order == order
	})

	if sIX != -1 {
		return &data[sIX], true
	} else {
		return &Section{}, false
	}
}

func FindNoteByItsOrder(notes []*Note, order int) *Note {
	noteIX := slices.IndexFunc(notes, func(note *Note) bool {
		return note.Order == order
	})

	if noteIX != -1 {
		return notes[noteIX]
	} else {
		return nil
	}
}

func FindNotesBySectionOrder(data ProgramModel, order int) ([]*Note, bool) {
	section, ok := FindSectionDataByOrder(data.SectionData, order)
	if !ok {
		return []*Note{}, false
	}
	notes, ok := data.UIControl.DisplayOrder[section.ID]

	return notes, ok

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

func (m ProgramModel) View() string {

	// The header
	allText := ""
	sectionLen := len(m.SectionData)

	dpo := m.UIControl.DisplayOrder
	sectionIDs := maps.Keys(dpo)

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
			sectionText = sectionHeaderStyle.Underline(true).
				// Italic(true).
				Background(lg.Color(ForegroundColor)).
				Foreground(lg.Color("#ffd700")).
				Render(section.Name)

		} else {
			sectionText = sectionHeaderStyle.Render(section.Name)
		}

		sectionText += "\n\n"
		sortedNotes := slices.Clone(notesInSection)
		slices.SortFunc(sortedNotes, func(a, b *Note) int {
			return a.Order - b.Order
		})

		// Iterate over our sortedNotes in the section
		for _, note := range sortedNotes {
			// Is the cursor pointing at this item?

			style := cardStyle
			style = style.Width((m.UIControl.TermSize.Width - 5) / (sectionLen + 3))
			// cursor := " " // no cursor
			if m.UIControl.RowCursor == note.Order && m.UIControl.SectionCursor == section.Order {
				// cursor = ">" // cursor!

				style = style.BorderStyle(lg.DoubleBorder()).Background(lg.Color(CardBackgroudColor))
			}

			// Is this item selected?
			// checked := " " // not selected
			if note.IsChecked {
				style = style.Background(lg.ANSIColor(7)).Foreground(lg.ANSIColor(8))
				// checked = "x"
			}

			// tmpS := fmt.Sprintf(" %s[%s] %s", cursor, checked, note.Content)
			tmpS := note.Content
			tmpS = style.Render(tmpS)
			sectionText += tmpS
			sectionText += "\n"
		}

		allText = lg.JoinHorizontal(lg.Top, allText, sectionText)
		if loopCnt < len(sectionList)-1 {
			allText = lg.JoinHorizontal(lg.Top, allText, "        ")
		}
	}

	// The footer

	if m.IsTextInputShown {
		allText += fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			m.InputPrompt,
			m.TextInput.View(),
			"(esc to cancel)\n",
		)
	} else {
		allText += "\nPress q to quit.\n"

	}

	allText += m.StatusText

	// DEBUG
	// allText += spew.Sdump(m.SectionData)
	allText += m.Debug

	// Send the UI for rendering

	return systemStyle.Width(m.UIControl.TermSize.Width - 3).Height(m.UIControl.TermSize.Height - 5).Render(allText)

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

func NewNote(content string, order int, sectionId int) *Note {
	return &Note{
		Content:     content,
		DateUpdated: time.Now(),
		DateCreated: time.Now(),
		Order:       order,
		SectionID:   sectionId,
	}
}

func NewSection(content string, order int, sectionId int) Section {
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
