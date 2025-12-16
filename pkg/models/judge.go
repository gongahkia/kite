package models

import "time"

// Judge represents a judge or judicial officer
type Judge struct {
	// Identifiers
	ID              string    `json:"id" validate:"required"`
	Name            string    `json:"name" validate:"required"`
	FullName        string    `json:"full_name,omitempty"`
	Title           string    `json:"title,omitempty"` // Justice, Judge, Magistrate, etc.

	// Court Information
	Court           string    `json:"court,omitempty"`
	Jurisdiction    string    `json:"jurisdiction,omitempty"`
	AppointmentDate *time.Time `json:"appointment_date,omitempty"`
	RetirementDate  *time.Time `json:"retirement_date,omitempty"`

	// Biography
	Education       []string  `json:"education,omitempty"`
	PreviousPositions []string `json:"previous_positions,omitempty"`
	Biography       string    `json:"biography,omitempty"`

	// Statistics
	CaseCount       int       `json:"case_count"`
	YearsOfService  int       `json:"years_of_service"`

	// Metadata
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at" validate:"required"`
	UpdatedAt       time.Time `json:"updated_at" validate:"required"`
}

// NewJudge creates a new Judge
func NewJudge(name string) *Judge {
	now := time.Now()
	return &Judge{
		Name:      name,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddCase increments the case count for the judge
func (j *Judge) AddCase() {
	j.CaseCount++
	j.UpdatedAt = time.Now()
}

// CalculateYearsOfService calculates years of service
func (j *Judge) CalculateYearsOfService() int {
	if j.AppointmentDate == nil {
		return 0
	}

	endDate := time.Now()
	if j.RetirementDate != nil {
		endDate = *j.RetirementDate
	}

	years := endDate.Year() - j.AppointmentDate.Year()
	if endDate.YearDay() < j.AppointmentDate.YearDay() {
		years--
	}

	j.YearsOfService = years
	return years
}
