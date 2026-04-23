package agent

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type XMLProtocol struct{}

func NewXMLProtocol() *XMLProtocol {
	return &XMLProtocol{}
}

// A2A Message types
const (
	IntentIntroduction   = "INTRODUCTION"
	IntentInterest      = "INTEREST"
	IntentNegotiation    = "NEGOTIATION"
	IntentOffer          = "OFFER"
	IntentAccept         = "ACCEPT"
	IntentDecline        = "DECLINE"
	IntentSchedule       = "SCHEDULE"
	IntentConfirm        = "CONFIRM"
	IntentWithdraw       = "WITHDRAW"
	IntentInquiry        = "INQUIRY"
)

type Message struct {
	XMLName   xml.Name `xml:"message"`
	Header    Header   `xml:"header"`
	Payload   Payload  `xml:"payload"`
}

type Header struct {
	MessageID  string `xml:"message_id"`
	Timestamp  string `xml:"timestamp"`
	SenderID   string `xml:"sender_id"`
	ReceiverID string `xml:"receiver_id"`
	ReplyTo    string `xml:"reply_to,omitempty"`
}

type Payload struct {
	Intent     string       `xml:"intent"`
	Parameters *Parameters  `xml:"parameters,omitempty"`
	Negotiation *Negotiation `xml:"negotiation,omitempty"`
}

type Parameters struct {
	Role         string `xml:"role,omitempty"`
	Name         string `xml:"name,omitempty"`
	Title        string `xml:"title,omitempty"`
	Company      string `xml:"company,omitempty"`
	Location     string `xml:"location,omitempty"`
	SalaryMin    int    `xml:"salary_min,omitempty"`
	SalaryMax    int    `xml:"salary_max,omitempty"`
	Experience   string `xml:"experience,omitempty"`
	Skills       string `xml:"skills,omitempty"`
	MatchID      string `xml:"match_id,omitempty"`
	InterestLevel string `xml:"interest_level,omitempty"`
}

type Negotiation struct {
	Round       int          `xml:"round,attr"`
	Type        string       `xml:"type,omitempty"`
	Current     int          `xml:"current,omitempty"`
	Target      int          `xml:"target,omitempty"`
	Currency    string       `xml:"currency,omitempty"`
	Concessions int          `xml:"concessions,omitempty"`
	Compensation *Compensation `xml:"compensation,omitempty"`
	StartDate   string       `xml:"start_date,omitempty"`
	Notes       string       `xml:"notes,omitempty"`
}

type Compensation struct {
	BaseSalary int    `xml:"base_salary"`
	Currency   string `xml:"currency"`
	Bonus      int    `xml:"bonus,omitempty"`
	Equity     string `xml:"equity,omitempty"`
}

func (xp *XMLProtocol) CreateMessage(senderID, receiverID, intent string, params *Parameters) *Message {
	return &Message{
		Header: Header{
			MessageID:  uuid.New().String(),
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			SenderID:   senderID,
			ReceiverID: receiverID,
		},
		Payload: Payload{
			Intent:     intent,
			Parameters: params,
		},
	}
}

func (xp *XMLProtocol) ParseMessage(xmlData []byte) (*Message, error) {
	var msg Message
	if err := xml.Unmarshal(xmlData, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	return &msg, nil
}

func (xp *XMLProtocol) SerializeMessage(msg *Message) ([]byte, error) {
	return xml.Marshal(msg)
}

func (xp *XMLProtocol) CreateIntroduction(senderID, receiverID, role, name string) *Message {
	return xp.CreateMessage(senderID, receiverID, IntentIntroduction, &Parameters{
		Role: role,
		Name: name,
	})
}

func (xp *XMLProtocol) CreateInterest(senderID, receiverID, matchID, interestLevel string) *Message {
	return xp.CreateMessage(senderID, receiverID, IntentInterest, &Parameters{
		MatchID:      matchID,
		InterestLevel: interestLevel,
	})
}

func (xp *XMLProtocol) CreateOffer(senderID, receiverID string, comp Compensation, startDate string) *Message {
	msg := xp.CreateMessage(senderID, receiverID, IntentOffer, nil)
	msg.Payload.Negotiation = &Negotiation{
		Round:        1,
		Compensation: &comp,
		StartDate:    startDate,
	}
	return msg
}

func (xp *XMLProtocol) CreateAccept(senderID, receiverID string) *Message {
	return xp.CreateMessage(senderID, receiverID, IntentAccept, nil)
}

func (xp *XMLProtocol) CreateDecline(senderID, receiverID, reason string) *Message {
	params := &Parameters{}
	if reason != "" {
		params.Skills = reason // Using Skills field for decline reason
	}
	return xp.CreateMessage(senderID, receiverID, IntentDecline, params)
}

func (xp *XMLProtocol) CreateSchedule(senderID, receiverID, matchID, format, datetime, location string) *Message {
	return xp.CreateMessage(senderID, receiverID, IntentSchedule, &Parameters{
		MatchID: matchID,
		Role:    format, // Reusing Role field for interview format
		Title:   datetime,
		Location: location,
	})
}

// XMLTimeFormat returns the timestamp format for XML messages
func XMLTimeFormat() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// GenerateXML serializes a message to XML bytes
func (xp *XMLProtocol) GenerateXML(msg *Message) ([]byte, error) {
	return xml.MarshalIndent(msg, "", "  ")
}
