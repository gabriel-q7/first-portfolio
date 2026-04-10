package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Project represents a portfolio project.
type Project struct {
	ID          uuid.UUID
	Name        string
	Description string
	Tech        []string
	URL         string
	Featured    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AIInsight holds AI-generated insight for a project.
type AIInsight struct {
	ProjectID   uuid.UUID
	Summary     string
	Tags        []string
	Confidence  float64
	GeneratedAt time.Time
}

// NewProject creates a new Project with a generated UUID and timestamps.
func NewProject(name, description string, tech []string, url string) *Project {
	now := time.Now().UTC()
	return &Project{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Tech:        tech,
		URL:         url,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks that required fields are populated.
func (p *Project) Validate() error {
	if p.Name == "" {
		return errors.New("project name is required")
	}
	if p.Description == "" {
		return errors.New("project description is required")
	}
	return nil
}

// Update mutates the project's mutable fields and bumps UpdatedAt.
func (p *Project) Update(name, desc string, tech []string, url string) {
	p.Name = name
	p.Description = desc
	p.Tech = tech
	p.URL = url
	p.UpdatedAt = time.Now().UTC()
}
