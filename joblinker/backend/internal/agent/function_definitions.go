package agent

// FunctionDefinition represents a tool/function that the agent can call
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GetFunctionDefinitions returns all available tool definitions for AI function calling
func GetFunctionDefinitions() []FunctionDefinition {
	return []FunctionDefinition{
		{
			Name:        "query_jobs",
			Description: "Search for jobs based on location, skills, salary range, and job type. Returns matching job listings with details.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "Desired work location (city, region, or 'remote')",
					},
					"skills": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Required skills for the job (e.g., ['python', 'django'])",
					},
					"salary_min": map[string]interface{}{
						"type":        "integer",
						"description": "Minimum annual salary in USD",
					},
					"job_type": map[string]interface{}{
						"type": "string",
						"enum": []string{"remote", "hybrid", "onsite"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results to return",
					},
				},
				"required": []string{"location"},
			},
		},
		{
			Name:        "get_candidate",
			Description: "Get detailed candidate profile including skills, experience, and preferences.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"candidate_id": map[string]interface{}{
						"type":        "string",
						"description": "Unique candidate identifier",
					},
				},
				"required": []string{"candidate_id"},
			},
		},
		{
			Name:        "create_offer",
			Description: "Create a job offer for a candidate match. Sends offer details to candidate.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"match_id": map[string]interface{}{
						"type":        "string",
						"description": "The match identifier linking candidate and job",
					},
					"salary": map[string]interface{}{
						"type":        "integer",
						"description": "Annual salary offer in USD",
					},
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Proposed start date (YYYY-MM-DD)",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Additional notes for the candidate",
					},
				},
				"required": []string{"match_id", "salary"},
			},
		},
		{
			Name:        "schedule_interview",
			Description: "Schedule an interview between the candidate and recruiter.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"match_id": map[string]interface{}{
						"type":        "string",
						"description": "The match identifier",
					},
					"datetime": map[string]interface{}{
						"type":        "string",
						"description": "Interview date and time (ISO 8601 format)",
					},
					"duration_minutes": map[string]interface{}{
						"type":        "integer",
						"default":     60,
						"description": "Interview duration in minutes",
					},
					"interview_type": map[string]interface{}{
						"type": "string",
						"enum": []string{"video", "phone", "onsite"},
					},
				},
				"required": []string{"match_id", "datetime"},
			},
		},
		{
			Name:        "search_candidates",
			Description: "Find candidates matching job requirements using vector similarity search.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"skills": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Required skills",
					},
					"location": map[string]interface{}{
						"type":        "string",
						"description": "Preferred location",
					},
					"experience_min": map[string]interface{}{
						"type":        "integer",
						"description": "Minimum years of experience",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum candidates to return",
					},
				},
				"required": []string{"skills"},
			},
		},
	}
}

// GetToolNames returns just the names of all available tools
func GetToolNames() []string {
	defs := GetFunctionDefinitions()
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = d.Name
	}
	return names
}

// GetToolByName returns a function definition by its name
func GetToolByName(name string) *FunctionDefinition {
	for _, def := range GetFunctionDefinitions() {
		if def.Name == name {
			return &def
		}
	}
	return nil
}