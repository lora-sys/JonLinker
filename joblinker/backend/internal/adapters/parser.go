package adapters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// ResumeParser handles parsing of resume files (PDF, DOCX, TXT).
type ResumeParser struct{}

// NewResumeParser creates a new ResumeParser.
func NewResumeParser() *ResumeParser {
	return &ResumeParser{}
}

// ParsedResume holds extracted resume data.
type ParsedResume struct {
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone"`
	Skills         []string `json:"skills"`
	Education      []string `json:"education"`
	Experience     []string `json:"experience"`
	Summary        string   `json:"summary"`
	RawText        string   `json:"raw_text"`
}

// ParseFile parses a resume from a file path.
func (p *ResumeParser) ParseFile(filePath string) (*ParsedResume, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	if len(content) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	return p.parseText(content), nil
}

// ParseFromURL downloads and parses a resume from a URL.
func (p *ResumeParser) ParseFromURL(url string) (*ParsedResume, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "pdf") {
		return p.parsePDFReader(resp.Body)
	}
	if strings.Contains(contentType, "word") || strings.Contains(contentType, "document") {
		return p.parseDOCXReader(resp.Body)
	}

	return p.parseTextReader(resp.Body)
}

func (p *ResumeParser) parseTextFile(filePath string) (*ParsedResume, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return p.parseText(string(data)), nil
}

func (p *ResumeParser) parseTextReader(r io.Reader) (*ParsedResume, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return p.parseText(string(data)), nil
}

func (p *ResumeParser) parsePDF(filePath string) *ParsedResume {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return p.parseText("")
	}
	return p.parsePDFBytes(data)
}

func (p *ResumeParser) parsePDFReader(data io.Reader) (*ParsedResume, error) {
	var buf strings.Builder
	_, err := io.Copy(&buf, data)
	if err != nil {
		return nil, err
	}
	return p.parsePDFString(buf.String()), nil
}

func (p *ResumeParser) parsePDFBytes(data []byte) *ParsedResume {
	return p.parsePDFString(string(data))
}

func (p *ResumeParser) parsePDFString(s string) *ParsedResume {
	text := extractPDFText(s)
	if text == "" {
		return p.parseText("")
	}
	return p.parseText(text)
}

func extractPDFText(s string) string {
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
		return nil, err
	}
	defer f.Close()
	return p.parseDOCXReader(f)
}

func (p *ResumeParser) parseDOCXReader(r io.Reader) (*ParsedResume, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	text := extractDOCXText(string(data))
	return p.parseText(text), nil
}

func extractDOCXText(s string) string {
	var result strings.Builder
	inTag := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			inTag = true
			continue
		}
		if s[i] == '>' {
			inTag = false
			result.WriteString(" ")
			continue
		}
		if !inTag {
			result.WriteByte(s[i])
		}
	}
	return cleanText(result.String())
}

func (p *ResumeParser) parseText(text string) *ParsedResume {
	resume := &ParsedResume{
		RawText: text,
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	resume.Name = extractName(lines)
	resume.Email = extractEmail(text)
	resume.Phone = extractPhone(text)
	resume.Skills = extractSkills(text)

	return resume
}

func cleanText(s string) string {
	result := strings.Builder{}
	inSpace := false
	for _, c := range s {
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			if !inSpace {
				result.WriteRune(' ')
				inSpace = true
			}
		} else {
			result.WriteRune(c)
			inSpace = false
		}
	}
	return strings.TrimSpace(result.String())
}

func extractEmail(s string) string {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	match := re.FindString(s)
	return match
}

func extractPhone(s string) string {
	re := regexp.MustCompile(`(\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	match := re.FindString(s)
	return match
}

func extractName(lines []string) string {
	if len(lines) > 0 && len(lines[0]) > 0 && !strings.Contains(lines[0], "@") && len(lines[0]) < 50 {
		return lines[0]
	}
	return ""
}

func extractSkills(s string) []string {
	skillKeywords := []string{
		"go", "golang", "python", "java", "javascript", "typescript", "rust",
		"react", "angular", "vue", "node.js", "node",
		"docker", "kubernetes", "k8s", "aws", "gcp", "azure",
		"sql", "postgresql", "mysql", "mongodb", "redis",
		"git", "ci/cd", "terraform", "ansible",
		"machine learning", "data science", "nlp",
		"agile", "scrum", "leadership",
	}

	var skills []string
	seen := make(map[string]bool)
	lower := strings.ToLower(s)
	for _, kw := range skillKeywords {
		if strings.Contains(lower, kw) && !seen[kw] {
			skills = append(skills, kw)
			seen[kw] = true
		}
	}
	return skills
}

// ToJSON serializes the parsed resume to JSON.
func (p *ParsedResume) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
