package agent

import (
	"time"
)

// Scheduler handles interview scheduling logic between agents
type Scheduler struct {
	timezone       string
	minNoticeDays  int
	maxAdvanceDays int
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		timezone:       "UTC",
		minNoticeDays:  1,
		maxAdvanceDays: 30,
	}
}

type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Available bool  `json:"available"`
}

type InterviewFormat string

const (
	FormatVideo  InterviewFormat = "video"
	FormatPhone  InterviewFormat = "phone"
	FormatOnsite InterviewFormat = "onsite"
)

// GenerateTimeSlots generates available interview time slots
func (s *Scheduler) GenerateTimeSlots(startDate time.Time, days int, durationMinutes int) []TimeSlot {
	var slots []TimeSlot

	for d := 0; d < days; d++ {
		date := startDate.AddDate(0, 0, d)
		// Business hours: 9 AM to 5 PM
		for hour := 9; hour < 17; hour++ {
			slotStart := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location())
			slotEnd := slotStart.Add(time.Duration(durationMinutes) * time.Minute)

			slots = append(slots, TimeSlot{
				Start:     slotStart,
				End:       slotEnd,
				Available: true,
			})
		}
	}

	return slots
}

// FindCommonSlots finds overlapping available times for two parties
func (s *Scheduler) FindCommonSlots(agent1Slots, agent2Slots []TimeSlot) []TimeSlot {
	slotMap := make(map[string]TimeSlot)

	for _, slot := range agent1Slots {
		if slot.Available {
			key := slot.Start.Format(time.RFC3339)
			slotMap[key] = slot
		}
	}

	var common []TimeSlot
	for _, slot := range agent2Slots {
		if !slot.Available {
			continue
		}
		key := slot.Start.Format(time.RFC3339)
		if _, exists := slotMap[key]; exists {
			common = append(common, slot)
		}
	}

	return common
}

// ProposeInterview creates an interview proposal with time and format
func (s *Scheduler) ProposeInterview(slot TimeSlot, format InterviewFormat, participants []string) *InterviewProposal {
	return &InterviewProposal{
		ProposedTime: slot.Start,
		Duration:     slot.End.Sub(slot.Start),
		Format:       format,
		Participants: participants,
		Status:       ProposalStatusPending,
	}
}

type InterviewProposal struct {
	ProposedTime time.Time
	Duration     time.Duration
	Format       InterviewFormat
	Participants []string
	Status       ProposalStatus
}

type ProposalStatus string

const (
	ProposalStatusPending   ProposalStatus = "pending"
	ProposalStatusConfirmed ProposalStatus = "confirmed"
	ProposalStatusDeclined  ProposalStatus = "declined"
	ProposalStatusRescheduled ProposalStatus = "rescheduled"
)

// IsWithinNoticePeriod checks if the proposal meets minimum notice requirement
func (s *Scheduler) IsWithinNoticePeriod(proposedTime time.Time) bool {
	noticeRequired := time.Now().Add(time.Duration(s.minNoticeDays) * 24 * time.Hour)
	return proposedTime.After(noticeRequired)
}

// IsTooFarAhead checks if the proposal is within maximum advance days
func (s *Scheduler) IsTooFarAhead(proposedTime time.Time) bool {
	maxAdvance := time.Now().Add(time.Duration(s.maxAdvanceDays) * 24 * time.Hour)
	return proposedTime.After(maxAdvance)
}

// FormatICSTime creates iCalendar formatted time string
func FormatICSTime(t time.Time) string {
	return t.Format("20060102T150405Z")
}

// GenerateICS generates an iCalendar event string
func GenerateICS(title, description string, start, end time.Time, location string) string {
	return `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//JobLinker//Interview//EN
BEGIN:VEVENT
UID:` + generateUID() + `
DTSTAMP:` + FormatICSTime(time.Now()) + `
DTSTART:` + FormatICSTime(start) + `
DTEND:` + FormatICSTime(end) + `
SUMMARY:` + title + `
DESCRIPTION:` + description + `
LOCATION:` + location + `
STATUS:CONFIRMED
END:VEVENT
END:VCALENDAR`
}

func generateUID() string {
	return time.Now().Format("20060102T150405") + "-joblinker@" + "scheduler"
}