package adapters

import (
	"context"
	"log"

	"joblinker/internal/core"
	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

// modelToCore converts model types to core types.

func modelJobToCore(j *model.Job) *core.Job {
	if j == nil {
		return nil
	}
	return &core.Job{
		ID:            j.ID,
		TenantID:      j.TenantID,
		CreatedAt:     j.CreatedAt,
		StructuredJSON: j.StructuredJSON,
	}
}

func modelAgentToCore(a *model.Agent) *core.Agent {
	if a == nil {
		return nil
	}
	return &core.Agent{
		ID:         a.ID,
		TenantID:   a.TenantID,
		UserID:     a.UserID,
		User:       modelUserToCore(a.User),
		ConfigJSON: a.ConfigJSON,
	}
}

func modelUserToCore(u *model.User) *core.User {
	if u == nil {
		return nil
	}
	return &core.User{
		ID:    u.ID,
		Email: u.Email,
	}
}

func modelOfferToCore(o *model.Offer) *core.Offer {
	if o == nil {
		return nil
	}
	return &core.Offer{
		ID:               o.ID,
		MatchID:          o.MatchID,
		TenantID:         o.TenantID,
		CompensationJSON: o.CompensationJSON,
		StartDate:        o.StartDate,
		Status:           core.OfferStatus(o.Status),
	}
}

func modelMatchToCore(m *model.Match) *core.Match {
	if m == nil {
		return nil
	}
	return &core.Match{
		ID:      m.ID,
		TenantID: m.TenantID,
		Job:     modelJobToCore(m.Job),
		Status:  core.MatchStatus(m.Status),
	}
}

func modelInterviewToCore(i *model.Interview) *core.Interview {
	if i == nil {
		return nil
	}
	return &core.Interview{
		ID:          i.ID,
		MatchID:     i.MatchID,
		TenantID:    i.TenantID,
		ScheduledAt: i.ScheduledAt,
		Format:      core.InterviewFormat(i.Format),
		Status:      core.InterviewStatus(i.Status),
	}
}

func coreToolCallToModel(t *core.AgentToolCall) *model.AgentToolCall {
	if t == nil {
		return nil
	}
	return &model.AgentToolCall{
		ID:        t.ID,
		MatchID:   t.MatchID,
		ToolName:  t.ToolName,
		Arguments: t.Arguments,
		Result:    t.Result,
		Status:    model.ToolStatus(t.Status),
	}
}

// CoreJobRepo wraps repository.JobRepository to satisfy core.JobRepository.
type CoreJobRepo struct {
	inner *repository.JobRepository
}

func NewCoreJobRepo(inner *repository.JobRepository) *CoreJobRepo {
	return &CoreJobRepo{inner: inner}
}

func (r *CoreJobRepo) Search(ctx context.Context, query string, skills []string, location string, salaryMin int, jobType string, limit int) ([]*core.Job, error) {
	jobs, err := r.inner.Search(ctx, query, skills, location, salaryMin, jobType, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*core.Job, len(jobs))
	for i, j := range jobs {
		result[i] = modelJobToCore(j)
	}
	return result, nil
}

// CoreAgentRepo wraps repository.AgentRepository to satisfy core.AgentRepository.
type CoreAgentRepo struct {
	inner *repository.AgentRepository
}

func NewCoreAgentRepo(inner *repository.AgentRepository) *CoreAgentRepo {
	return &CoreAgentRepo{inner: inner}
}

func (r *CoreAgentRepo) GetByID(id uuid.UUID) (*core.Agent, error) {
	agent, err := r.inner.GetByID(id)
	if err != nil {
		return nil, err
	}
	return modelAgentToCore(agent), nil
}

func (r *CoreAgentRepo) SearchBySkills(ctx context.Context, skills []string, location string, experienceMin int, limit int) ([]*core.Agent, error) {
	agents, err := r.inner.SearchBySkills(ctx, skills, location, experienceMin, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*core.Agent, len(agents))
	for i, a := range agents {
		result[i] = modelAgentToCore(a)
	}
	return result, nil
}

// CoreMatchRepo wraps repository.MatchRepository to satisfy core.MatchRepository.
type CoreMatchRepo struct {
	inner *repository.MatchRepository
}

func NewCoreMatchRepo(inner *repository.MatchRepository) *CoreMatchRepo {
	return &CoreMatchRepo{inner: inner}
}

func (r *CoreMatchRepo) GetByID(id uuid.UUID) (*core.Match, error) {
	match, err := r.inner.GetByID(id)
	if err != nil {
		return nil, err
	}
	return modelMatchToCore(match), nil
}

func (r *CoreMatchRepo) UpdateStatus(_ context.Context, id uuid.UUID, status core.MatchStatus) error {
	return r.inner.UpdateStatus(id, model.MatchStatus(status))
}

func (r *CoreMatchRepo) CreateToolCall(ctx context.Context, toolCall *core.AgentToolCall) error {
	return r.inner.CreateToolCall(ctx, coreToolCallToModel(toolCall))
}

// CoreOfferRepo wraps repository.OfferRepository to satisfy core.OfferRepository.
type CoreOfferRepo struct {
	inner *repository.OfferRepository
}

func NewCoreOfferRepo(inner *repository.OfferRepository) *CoreOfferRepo {
	return &CoreOfferRepo{inner: inner}
}

func (r *CoreOfferRepo) Create(offer *core.Offer) error {
	// Build model offer from core
	modelOffer := &model.Offer{
		ID:               offer.ID,
		MatchID:          offer.MatchID,
		TenantID:         offer.TenantID,
		CompensationJSON: offer.CompensationJSON,
		StartDate:        offer.StartDate,
		Status:           model.OfferStatus(offer.Status),
	}
	return r.inner.Create(modelOffer)
}

// CoreInterviewRepo wraps repository.InterviewRepository to satisfy core.InterviewRepository.
type CoreInterviewRepo struct {
	inner *repository.InterviewRepository
}

func NewCoreInterviewRepo(inner *repository.InterviewRepository) *CoreInterviewRepo {
	return &CoreInterviewRepo{inner: inner}
}

func (r *CoreInterviewRepo) Create(interview *core.Interview) error {
	modelInterview := &model.Interview{
		ID:          interview.ID,
		MatchID:     interview.MatchID,
		TenantID:    interview.TenantID,
		ScheduledAt: interview.ScheduledAt,
		Format:      model.InterviewFormat(interview.Format),
		Status:      model.InterviewStatus(interview.Status),
	}
	return r.inner.Create(modelInterview)
}

// Ensure adapters implement the core interfaces.
var _ core.JobRepository = (*CoreJobRepo)(nil)
var _ core.AgentRepository = (*CoreAgentRepo)(nil)
var _ core.MatchRepository = (*CoreMatchRepo)(nil)
var _ core.OfferRepository = (*CoreOfferRepo)(nil)
var _ core.InterviewRepository = (*CoreInterviewRepo)(nil)

// init log to confirm compilation-time checks
func init() {
	log.Println("[adapters] Core repository wrappers compiled successfully")
}
