package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ResumeParser struct{}

func NewResumeParser() *ResumeParser {
	return &ResumeParser{}
}

type ParsedResume struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone"`
	Skills       []string `json:"skills"`
	Experience   []string `json:"experience"`
	Education    []string `json:"education"`
	Summary      string   `json:"summary"`
	RawText      string   `json:"raw_text"`
}

// ParseFile extracts structured data from a resume file
func (p *ResumeParser) ParseFile(filePath string) (*ParsedResume, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".txt":
		return p.parseTextFile(filePath)
	case ".pdf":
		return p.parsePDF(filePath)
	case ".docx":
		return p.parseDOCX(filePath)
	case ".doc":
		return p.parseDOC(filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// ParseFromURL downloads and parses a resume from URL
func (p *ResumeParser) ParseFromURL(url string) (*ParsedResume, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	// Determine file type from Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "pdf") {
		return p.parsePDFReader(resp.Body)
	}
	if strings.Contains(contentType, "word") || strings.Contains(contentType, "document") {
		return p.parseDOCXReader(resp.Body)
	}

	// Default to text
	return p.parseTextReader(resp.Body)
}

func (p *ResumeParser) parseTextFile(filePath string) (*ParsedResume, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return p.parseText(string(data)), nil
}

func (p *ResumeParser) parseTextReader(r io.Reader) (*ParsedResume, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	return p.parseText(string(data)), nil
}

func (p *ResumeParser) parsePDF(filePath string) (*ParsedResume, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	return p.parsePDFReader(f)
}

func (p *ResumeParser) parsePDFReader(r io.Reader) (*ParsedResume, error) {
	// For production, use a PDF parsing library like pdfcpu or unipdf
	// This is a simplified version that extracts text patterns
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Extract text between BT (Begin Text) and ET (End Text) markers
	text := extractPDFText(string(data))
	return p.parseText(text), nil
}

func extractPDFText(s string) string {
	// Simplified PDF text extraction
	var result strings.Builder
	inText := false
	for i := 0; i < len(s); i++ {
		if i+1 < len(s) && s[i] == 'B' && s[i+1] == 'T' {
			inText = true
			i++
			continue
		}
		if i+1 < len(s) && s[i] == 'E' && s[i+1] == 'T' {
			inText = false
			result.WriteString(" ")
			i++
			continue
		}
		if inText {
			result.WriteByte(s[i])
		}
	}
	return cleanText(result.String())
}

func (p *ResumeParser) parseDOCX(filePath string) (*ParsedResume, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	return p.parseDOCXReader(f)
}

func (p *ResumeParser) parseDOCXReader(r io.Reader) (*ParsedResume, error) {
	// For production, use a library like unioffice or docx
	// DOCX is a ZIP file containing XML - simplified extraction
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Look for <w:t> tags which contain text in DOCX
	text := extractDOCXText(string(data))
	return p.parseText(text), nil
}

func extractDOCXText(s string) string {
	var result strings.Builder
	// Extract text between <w:t> tags
	lines := strings.Split(s, "<w:t")
	for i := 1; i < len(lines); i++ {
		part := lines[i]
		end := strings.Index(part, "</w:t>")
		if end > 0 {
			text := part[:end]
			// Skip to after >
			if idx := strings.Index(text, ">"); idx >= 0 {
				text = text[idx+1:]
			}
			result.WriteString(text)
			result.WriteString(" ")
		}
	}
	return cleanText(result.String())
}

func (p *ResumeParser) parseDOC(filePath string) (*ParsedResume, error) {
	// Old DOC format - in production use a library like antiword
	// For now, treat as text
	return p.parseTextFile(filePath)
}

func (p *ResumeParser) parseText(text string) *ParsedResume {
	cleaned := cleanText(text)
	lines := strings.Split(cleaned, "\n")

	resume := &ParsedResume{
		Skills:     []string{},
		Experience: []string{},
		Education:  []string{},
		RawText:    cleaned,
	}

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)

		// Detect sections
		if strings.Contains(lower, "experience") || strings.Contains(lower, "work history") {
			currentSection = "experience"
			continue
		}
		if strings.Contains(lower, "education") || strings.Contains(lower, "academic") {
			currentSection = "education"
			continue
		}
		if strings.Contains(lower, "skill") {
			currentSection = "skills"
			continue
		}
		if strings.Contains(lower, "summary") || strings.Contains(lower, "objective") {
			currentSection = "summary"
			continue
		}

		// Extract data based on section or patterns
		switch currentSection {
		case "experience":
			if line != "" {
				resume.Experience = append(resume.Experience, line)
			}
		case "education":
			if line != "" {
				resume.Education = append(resume.Education, line)
			}
		case "skills":
			skills := extractSkills(line)
			resume.Skills = append(resume.Skills, skills...)
		case "summary":
			resume.Summary += line + " "
		default:
			// Try to extract name, email, phone from early lines
			if resume.Name == "" && isName(line) {
				resume.Name = line
			}
			if resume.Email == "" {
				if email := extractEmail(line); email != "" {
					resume.Email = email
				}
			}
			if resume.Phone == "" {
				if phone := extractPhone(line); phone != "" {
					resume.Phone = phone
				}
			}
			// Also extract skills anywhere
			skills := extractSkills(line)
			resume.Skills = append(resume.Skills, skills...)
		}
	}

	// Deduplicate skills
	seen := make(map[string]bool)
	unique := []string{}
	for _, s := range resume.Skills {
		lower := strings.ToLower(strings.TrimSpace(s))
		if !seen[lower] && len(lower) > 1 {
			seen[lower] = true
			unique = append(unique, strings.TrimSpace(s))
		}
	}
	resume.Skills = unique

	return resume
}

func cleanText(s string) string {
	// Remove excessive whitespace
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

func extractEmail(s string) string {
	// Simple email pattern
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			// Search backwards
			start := i
			for start > 0 && isEmailChar(s[start-1]) {
				start--
			}
			// Search forwards
			end := i
			for end < len(s) && isEmailChar(s[end]) {
				end++
			}
			if start < i && end > i+1 {
				return s[start:end]
			}
		}
	}
	return ""
}

func isEmailChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '@' || c == '.' || c == '_' || c == '-'
}

func extractPhone(s string) string {
	// Simple phone pattern - digits with optional separators
	var digits []byte
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			digits = append(digits, s[i])
		}
	}
	if len(digits) >= 10 {
		return string(digits)
	}
	return ""
}

func isName(s string) bool {
	// Simple heuristic: 2-4 words, each starting with uppercase
	words := strings.Split(s, " ")
	if len(words) < 2 || len(words) > 4 {
		return false
	}
	count := 0
	for _, w := range words {
		if len(w) > 0 && w[0] >= 'A' && w[0] <= 'Z' {
			count++
		}
	}
	return count >= 2
}

func extractSkills(s string) []string {
	// Common skill keywords to look for
	skillKeywords := []string{
		"Go", "Golang", "Python", "Java", "JavaScript", "TypeScript", "C++", "C#",
		"React", "Angular", "Vue", "Node.js", "Next.js",
		"PostgreSQL", "MySQL", "MongoDB", "Redis", "Elasticsearch",
		"AWS", "Azure", "GCP", "Docker", "Kubernetes",
		"Git", "Linux", "REST", "GraphQL", "API",
		"Machine Learning", "AI", "Data Science", "SQL",
		"HTML", "CSS", "Tailwind", "Bootstrap",
		"Agile", "Scrum", "Project Management",
	}

	s = strings.ToUpper(s)
	var skills []string
	for _, skill := range skillKeywords {
		if strings.Contains(s, strings.ToUpper(skill)) {
			skills = append(skills, skill)
		}
	}
	return skills
}

// ToJSON converts parsed resume to JSON string
func (p *ParsedResume) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
