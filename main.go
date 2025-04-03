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
			case "alt+up":
				{
					if m.UIControl.RowCursor == 0 {
						break
					}

					notes, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor)
					if !ok {
						break
					}

					curNote := FindNoteByItsOrder(notes, m.UIControl.RowCursor)
					passNote := FindNoteByItsOrder(notes, m.UIControl.RowCursor-1)

					curNote.Order = curNote.Order - 1
					passNote.Order = passNote.ID + 1
					m.UIControl.RowCursor--

				}
			case "alt+down":
				{
					notes, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor)
					if !ok {
						break
					}
					if m.UIControl.RowCursor == len(notes)-1 {
						break
					}

					curNote := FindNoteByItsOrder(notes, m.UIControl.RowCursor)
					nextNote := FindNoteByItsOrder(notes, m.UIControl.RowCursor+1)

					curNote.Order++
					nextNote.Order--
					m.UIControl.RowCursor++
				}
			case "alt+left":
				{
					if m.UIControl.SectionCursor == 0 {
						break
					}

					currNote := FindNoteByBothOrder(m, m.UIControl.SectionCursor, m.UIControl.RowCursor)
					if currNote == nil {
						break
					}

					section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor-1)
					if !ok {
						break
					}
					currNote.SectionID = section.ID
					currNote.DateUpdated = time.Now()

					passSec, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor-1)
					if !ok {
						break
					}
					passSec = append(passSec, currNote)
					RecalulateNoteOrder(passSec)

					curSec, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor)
					if !ok {
						break
					}
					curSec = slices.DeleteFunc(curSec, func(n *Note) bool {
						return n.Order == currNote.Order
					})
					RecalulateNoteOrder(curSec)

					m.RepopulateDisplayOrder()
					m.UIControl.SectionCursor--

				}

			case "alt+right":
				{
					if m.UIControl.SectionCursor == len(m.SectionData)-1 {
						break
					}

					currNote := FindNoteByBothOrder(m, m.UIControl.SectionCursor, m.UIControl.RowCursor)
					if currNote == nil {
						break
					}

					section, ok := FindSectionDataByOrder(m.SectionData, m.UIControl.SectionCursor+1)
					if !ok {
						break
					}
					currNote.SectionID = section.ID
					currNote.DateUpdated = time.Now()

					nextSec, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor+1)
					if !ok {
						break
					}
					nextSec = append(nextSec, currNote)
					RecalulateNoteOrder(nextSec)

					curSec, ok := FindNotesBySectionOrder(m, m.UIControl.SectionCursor)
					if !ok {
						break
					}
					curSec = slices.DeleteFunc(curSec, func(n *Note) bool {
						return n.Order == currNote.Order
					})
					RecalulateNoteOrder(curSec)

					m.UIControl.SectionCursor++
				}

				// default:
				// 	{
				// 		m.Debug = msg.String()
				// 	}
			}
		}
	}

	m.RepopulateDisplayOrder()
	return m, cmd
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
