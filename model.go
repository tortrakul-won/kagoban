package main

import (
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

type ProgramModel struct {
	UIControl        UIControl
	Notes            []*Note
	SectionData      []Section
	IsInit           bool
	IsTextInputShown bool
	TextInput        textinput.Model
	InputPrompt      string
	Operation        string // Might change to enum
	Debug            string
	StatusText       string
}

func initialModel() ProgramModel {

	model, err := LoadProgramStateFromJson()
	if os.IsNotExist(err) {
		LoadBlankProgramState()
	} else if err != nil {
		model = LoadMockData()
	}
	ti := NewTextInputSetting()

	model.TextInput = ti
	return model
}

func LoadProgramStateFromJson() (ProgramModel, error) {
	jsonData, err := os.ReadFile("./data/save_file.json")
	if err != nil {
		return ProgramModel{}, err
	}
	type StateDTO struct {
		SectionData []Section
		Notes       []*Note
	}

	var dto StateDTO
	if err := json.Unmarshal(jsonData, &dto); err != nil {
		return ProgramModel{}, err
	}

	return ProgramModel{Notes: dto.Notes, SectionData: dto.SectionData}, nil
}

func LoadMockData() ProgramModel {
	mockNotes := []*Note{} // Create a slice with capacity for 4 items
	for i := range 4 {
		mockNotes = append(mockNotes, NewNote("test"+strconv.Itoa(i), i, 0))
	}
	for i := range 2 {
		mockNotes = append(mockNotes, NewNote("test"+strconv.Itoa(i), i, 1))
	}

	return ProgramModel{
		Notes: mockNotes,
		SectionData: []Section{
			{ID: 0, Order: 0, Name: "Uncategorized"},
			{ID: 1, Order: 1, Name: "Inbox"},
		},
	}
}

func LoadBlankProgramState() ProgramModel {

	return ProgramModel{
		Notes: []*Note{},
		SectionData: []Section{
			{ID: 0, Order: 0, Name: "Inbox"},
		},
	}
}

func (m *ProgramModel) RepopulateDisplayOrder() {
	m.UIControl.DisplayOrder = make(map[int][]*Note)
	dp := m.UIControl.DisplayOrder

	for _, section := range m.SectionData {
		dp[section.ID] = make([]*Note, 0)
	}
	for _, notePtr := range m.Notes {
		dp[notePtr.SectionID] = append(dp[notePtr.SectionID], notePtr)
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
