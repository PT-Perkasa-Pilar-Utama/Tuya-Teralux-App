package dtos

// CanonicalMeetingSummary defines the provider-agnostic contract for meeting summary results.
// All AI providers (Gemini, OpenAI, Groq, Orion) must produce output that can be normalized
// into this structure. The backend generates final markdown from this validated struct.
type CanonicalMeetingSummary struct {
	Metadata                SummaryMetadata         `json:"metadata"`
	Agenda                  string                  `json:"agenda,omitempty"`
	BackgroundAndObjective  string                  `json:"background_and_objective,omitempty"`
	MainDiscussionSections  []DiscussionSection     `json:"main_discussion_sections,omitempty"`
	RolesAndResponsibilities []RoleResponsibility   `json:"roles_and_responsibilities,omitempty"`
	ActionItems             []ActionItem            `json:"action_items,omitempty"`
	DecisionsMade           []Decision              `json:"decisions_made,omitempty"`
	OpenIssues              []OpenIssue             `json:"open_issues,omitempty"`
	RisksAndMitigation      []Risk                  `json:"risks_and_mitigation,omitempty"`
	AdditionalNotes         string                  `json:"additional_notes,omitempty"`
}

// SummaryMetadata contains meeting metadata fields that are required for every summary.
type SummaryMetadata struct {
	MeetingTitle  string   `json:"meeting_title"`
	Date          string   `json:"date,omitempty"`
	Location      string   `json:"location,omitempty"`
	Participants  []string `json:"participants,omitempty"`
	Context       string   `json:"context,omitempty"`
	Style         string   `json:"style,omitempty"` // e.g., "minutes", "executive"
	Language      string   `json:"language"`        // Required: output language of the summary
}

// DiscussionSection represents a main discussion topic with its key points.
type DiscussionSection struct {
	Title       string          `json:"title"`
	KeyPoints   []DiscussionPoint `json:"key_points"`
	Decisions   []string        `json:"decisions,omitempty"`
	ActionItems []string        `json:"action_items,omitempty"`
}

// DiscussionPoint represents a single point within a discussion section.
type DiscussionPoint struct {
	Content   string `json:"content"`
	Speaker   string `json:"speaker,omitempty"`
	Timestamp string `json:"timestamp,omitempty"` // Optional timestamp reference
}

// RoleResponsibility represents a role or responsibility assignment from the meeting.
type RoleResponsibility struct {
	Role        string `json:"role"`
	AssignedTo  string `json:"assigned_to,omitempty"`
	Description string `json:"description,omitempty"`
}
