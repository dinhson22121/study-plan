// Package domain defines the curriculum bounded context: the Subject → Chapter
// → Topic catalog hierarchy and its repository port. This context is the source
// of truth for topic identity that question and content reference.
package domain

import (
	"strings"

	shared "github.com/son-ngo/edu-app/internal/shared/domain"
)

const (
	minGrade = 1
	maxGrade = 12
)

// Subject is a top-level catalog entry (e.g. MATH, ENGLISH, PHYSICS).
type Subject struct {
	ID         string
	Code       string
	Name       string
	GradeLevel int
}

// NewSubject validates and constructs a subject.
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

// Chapter groups topics within a subject, ordered by OrderIndex.
type Chapter struct {
	ID         string
	SubjectID  string
	Title      string
	OrderIndex int
}

// NewChapter validates and constructs a chapter.
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

// Topic is a leaf learning unit within a chapter (e.g. "Logarit").
type Topic struct {
	ID         string
	ChapterID  string
	Title      string
	OrderIndex int
}

// NewTopic validates and constructs a topic.
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
