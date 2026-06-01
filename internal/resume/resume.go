package resume

import "encoding/json"

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

func (p *CandidateProfile) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if v, ok := m["name"]; ok {
		json.Unmarshal(v, &p.Name)
	}
	if v, ok := m["title"]; ok {
		json.Unmarshal(v, &p.Title)
	}
	if v, ok := m["skills"]; ok {
		json.Unmarshal(v, &p.Skills)
	}
	if v, ok := m["experience"]; ok {
		json.Unmarshal(v, &p.Experience)
	}
	if v, ok := m["phone"]; ok {
		json.Unmarshal(v, &p.Phone)
	}
	if v, ok := m["email"]; ok {
		json.Unmarshal(v, &p.Email)
	}
	if v, ok := m["summary"]; ok {
		json.Unmarshal(v, &p.Summary)
	}
	if v, ok := m["hobbies"]; ok {
		json.Unmarshal(v, &p.Hobbies)
	}
	if v, ok := m["education"]; ok {
		var arr []Education
		if err := json.Unmarshal(v, &arr); err == nil {
			p.Education = arr
		} else {
			var single Education
			if err := json.Unmarshal(v, &single); err == nil {
				p.Education = []Education{single}
			}
		}
	}
	return nil
}

type Experience struct {
	Company     string `json:"company"`
	Title       string `json:"title"`
	Duration    string `json:"duration"`
	Description string `json:"description"`
}

func (e *Experience) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if v, ok := m["company"]; ok {
		json.Unmarshal(v, &e.Company)
	}
	if v, ok := m["title"]; ok {
		json.Unmarshal(v, &e.Title)
	}
	if v, ok := m["position"]; ok {
		json.Unmarshal(v, &e.Title)
	}
	if v, ok := m["duration"]; ok {
		json.Unmarshal(v, &e.Duration)
	}
	if v, ok := m["description"]; ok {
		json.Unmarshal(v, &e.Description)
	}
	return nil
}

type Education struct {
	School   string `json:"school"`
	Degree   string `json:"degree"`
	Major    string `json:"major"`
	Duration string `json:"duration"`
}
