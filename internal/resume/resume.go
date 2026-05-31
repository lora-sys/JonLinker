package resume

type CandidateProfile struct {
	Name       string       `json:"name"`
	Title      string       `json:"title"`
	Skills     []string     `json:"skills"`
	Experience []Experience `json:"experience"`
	Education  []Education  `json:"education"`
	Phone      string       `json:"phone"`
	Email      string       `json:"email"`
	Summary    string       `json:"summary"`
	Hobbies    []string     `json:"hobbies"`
}

type Experience struct {
	Company     string `json:"company"`
	Title       string `json:"title"`
	Duration    string `json:"duration"`
	Description string `json:"description"`
}

type Education struct {
	School   string `json:"school"`
	Degree   string `json:"degree"`
	Major    string `json:"major"`
	Duration string `json:"duration"`
}
