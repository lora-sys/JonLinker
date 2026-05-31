package query_jobs

import (
	"testing"
)

func TestParseIndeedJobs(t *testing.T) {
	md := `- |     |
| --- |
| ### [前端开发工程师](https://cn.indeed.com/jobs?jk=123)<br>字节跳动<br>北京市 |
前瞻性，能够结合业务场景设计 AI 技术方案并推动落地。
- |     |
| --- |
| ### [Go后端开发](https://cn.indeed.com/jobs?jk=456)<br>阿里巴巴<br>杭州市 |
1. 负责后端服务开发
2. 系统架构设计
- |     |
| --- |
| ### [产品经理](https://cn.indeed.com/jobs?jk=789)<br>腾讯<br>深圳市 |
`

	jobs := parseIndeedJobs(md)

	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d: %+v", len(jobs), jobs)
	}

	tests := []struct {
		idx     int
		title   string
		company string
	}{
		{0, "前端开发工程师", "字节跳动"},
		{1, "Go后端开发", "阿里巴巴"},
		{2, "产品经理", "腾讯"},
	}

	for _, tt := range tests {
		job := jobs[tt.idx]
		if job.Title != tt.title {
			t.Errorf("job[%d].Title = %q, want %q", tt.idx, job.Title, tt.title)
		}
		if job.Company != tt.company {
			t.Errorf("job[%d].Company = %q, want %q", tt.idx, job.Company, tt.company)
		}
	}
}

func TestParseIndeedJobs_Empty(t *testing.T) {
	jobs := parseIndeedJobs("")
	if len(jobs) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestParseIndeedJobs_NoJobs(t *testing.T) {
	md := `没有找到相关职位。
试试其他关键词吧。`
	jobs := parseIndeedJobs(md)
	if len(jobs) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(jobs))
	}
}
