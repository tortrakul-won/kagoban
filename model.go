package main

import (
	"encoding/json"
	"os"
	"strconv"

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
