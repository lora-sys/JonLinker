package job

type Job struct {
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Location    string   `json:"location"`
	Salary      string   `json:"salary"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Source      string   `json:"source"`
}

type RankedJob struct {
	Job
	MatchScore int      `json:"match_score"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

type SearchResponse struct {
	Message     string       `json:"message,omitempty"`
	Jobs        []RankedJob  `json:"jobs,omitempty"`
	Application *Application `json:"application,omitempty"`
}

type Application struct {
	JobTitle    string   `json:"job_title"`
	Company     string   `json:"company"`
	CoverLetter string   `json:"cover_letter"`
	ResumeMD    string   `json:"resume_md"`
	Highlights  []string `json:"highlights"`
	GeneratedAt string   `json:"generated_at"`
}

type ApplyRequest struct {
	JobURL    string `json:"job_url"`
	SessionID string `json:"session_id"`
}

type UploadResponse struct {
	SessionID string `json:"session_id"`
	Text      string `json:"text"`
}
