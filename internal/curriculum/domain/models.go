package domain

import (
	"strings"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const (
	minGrade = 1
	maxGrade = 12
)

type Subject struct {
	ID         string
	Code       string
	Name       string
	GradeLevel int
}

func NewSubject(id, code, name string, gradeLevel int) (*Subject, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, shared.ErrValidation.WithMessage("subject code is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, shared.ErrValidation.WithMessage("subject name is required")
	}
	if gradeLevel < minGrade || gradeLevel > maxGrade {
		return nil, shared.ErrValidation.WithMessage("grade level must be between 1 and 12")
	}
	return &Subject{ID: id, Code: code, Name: strings.TrimSpace(name), GradeLevel: gradeLevel}, nil
}

type Chapter struct {
	ID         string
	SubjectID  string
	Title      string
	OrderIndex int
}

func NewChapter(id, subjectID, title string, orderIndex int) (*Chapter, error) {
	if subjectID == "" {
		return nil, shared.ErrValidation.WithMessage("subject id is required")
	}
	if strings.TrimSpace(title) == "" {
		return nil, shared.ErrValidation.WithMessage("chapter title is required")
	}
	if orderIndex < 0 {
		return nil, shared.ErrValidation.WithMessage("order index must be non-negative")
	}
	return &Chapter{ID: id, SubjectID: subjectID, Title: strings.TrimSpace(title), OrderIndex: orderIndex}, nil
}

type Topic struct {
	ID         string
	ChapterID  string
	Title      string
	OrderIndex int
}

func NewTopic(id, chapterID, title string, orderIndex int) (*Topic, error) {
	if chapterID == "" {
		return nil, shared.ErrValidation.WithMessage("chapter id is required")
	}
	if strings.TrimSpace(title) == "" {
		return nil, shared.ErrValidation.WithMessage("topic title is required")
	}
	if orderIndex < 0 {
		return nil, shared.ErrValidation.WithMessage("order index must be non-negative")
	}
	return &Topic{ID: id, ChapterID: chapterID, Title: strings.TrimSpace(title), OrderIndex: orderIndex}, nil
}
