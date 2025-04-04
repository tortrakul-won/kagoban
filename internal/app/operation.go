package app

import (
	"slices"
)

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
