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

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Jobs   []RankedJob `json:"jobs"`
	Intent UserIntent  `json:"intent"`
}

type UserIntent struct {
	Keyword    string `json:"keyword"`
	City       string `json:"city"`
	SalaryMin  int    `json:"salary_min"`
	Experience string `json:"experience"`
}
