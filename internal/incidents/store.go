package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Status string

const (
	StatusOpen          Status = "open"
	StatusInvestigating Status = "investigating"
	StatusMitigated     Status = "mitigated"
	StatusResolved      Status = "resolved"
)

type Incident struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Service     string    `json:"service"`
	Severity    Severity  `json:"severity"`
	Status      Status    `json:"status"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	Title       string   `json:"title"`
	Service     string   `json:"service"`
	Severity    Severity `json:"severity"`
	Status      Status   `json:"status"`
	Description string   `json:"description"`
	Owner       string   `json:"owner"`
}

type UpdateInput struct {
	Title       string   `json:"title"`
	Service     string   `json:"service"`
	Severity    Severity `json:"severity"`
	Status      Status   `json:"status"`
	Description string   `json:"description"`
	Owner       string   `json:"owner"`
}

var ErrNotFound = errors.New("incident not found")

type ListOptions struct {
	Service  string
	Status   Status
	Severity Severity
	Owner    string
	Search   string
	Limit    int
	Offset   int
}

type ListResult struct {
	Items []Incident
	Total int
}

type Stats struct {
	Total      int              `json:"total"`
	ByStatus   map[Status]int   `json:"by_status"`
	BySeverity map[Severity]int `json:"by_severity"`
	ByService  map[string]int   `json:"by_service"`
}

type Store interface {
	List(context.Context, ListOptions) (ListResult, error)
	Get(context.Context, int64) (Incident, error)
	Stats(context.Context) (Stats, error)
	Create(context.Context, CreateInput) (Incident, error)
	Update(context.Context, int64, UpdateInput) (Incident, error)
	Delete(context.Context, int64) error
}

type FileStore struct {
	mu    sync.RWMutex
	seq   atomic.Int64
	path  string
	items []Incident
}

func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{path: path}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *FileStore) Seed(seed []Incident) error {
	s.mu.RLock()
	empty := len(s.items) == 0
	s.mu.RUnlock()
	if !empty {
		return nil
	}

	for _, item := range seed {
		if _, err := s.Create(context.Background(), CreateInput{
			Title:       item.Title,
			Service:     item.Service,
			Severity:    item.Severity,
			Status:      item.Status,
			Description: item.Description,
			Owner:       item.Owner,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStore) List(_ context.Context, options ListOptions) (ListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := slices.Clone(s.items)
	slices.Reverse(items)
	filtered := make([]Incident, 0, len(items))
	for _, item := range items {
		if !matches(item, options) {
			continue
		}
		filtered = append(filtered, item)
	}

	total := len(filtered)
	start := clampNonNegative(options.Offset)
	if start > total {
		start = total
	}
	end := total
	if options.Limit > 0 && start+options.Limit < end {
		end = start + options.Limit
	}

	return ListResult{
		Items: filtered[start:end],
		Total: total,
	}, nil
}

func (s *FileStore) Create(_ context.Context, input CreateInput) (Incident, error) {
	if err := validate(input.Title, input.Service, input.Severity, input.Status); err != nil {
		return Incident{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	item := Incident{
		ID:          s.seq.Add(1),
		Title:       strings.TrimSpace(input.Title),
		Service:     strings.TrimSpace(input.Service),
		Severity:    input.Severity,
		Status:      input.Status,
		Description: strings.TrimSpace(input.Description),
		Owner:       strings.TrimSpace(input.Owner),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.items = append(s.items, item)
	if err := s.persist(); err != nil {
		s.items = s.items[:len(s.items)-1]
		s.seq.Add(-1)
		return Incident{}, err
	}
	return item, nil
}

func (s *FileStore) Get(_ context.Context, id int64) (Incident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}

	return Incident{}, ErrNotFound
}

func (s *FileStore) Update(_ context.Context, id int64, input UpdateInput) (Incident, error) {
	if err := validate(input.Title, input.Service, input.Severity, input.Status); err != nil {
		return Incident{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.items {
		if s.items[index].ID != id {
			continue
		}

		original := s.items[index]
		item := &s.items[index]
		item.Title = strings.TrimSpace(input.Title)
		item.Service = strings.TrimSpace(input.Service)
		item.Severity = input.Severity
		item.Status = input.Status
		item.Description = strings.TrimSpace(input.Description)
		item.Owner = strings.TrimSpace(input.Owner)
		item.UpdatedAt = time.Now().UTC()
		if err := s.persist(); err != nil {
			s.items[index] = original
			return Incident{}, err
		}
		return *item, nil
	}

	return Incident{}, ErrNotFound
}

func (s *FileStore) Stats(_ context.Context) (Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := Stats{
		Total:      len(s.items),
		ByStatus:   make(map[Status]int),
		BySeverity: make(map[Severity]int),
		ByService:  make(map[string]int),
	}

	for _, item := range s.items {
		stats.ByStatus[item.Status]++
		stats.BySeverity[item.Severity]++
		stats.ByService[item.Service]++
	}

	return stats, nil
}

func (s *FileStore) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.items {
		if s.items[index].ID != id {
			continue
		}

		s.items = append(s.items[:index], s.items[index+1:]...)
		return s.persist()
	}

	return ErrNotFound
}

func (s *FileStore) load() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &s.items); err != nil {
		return err
	}
	for _, item := range s.items {
		if item.ID > s.seq.Load() {
			s.seq.Store(item.ID)
		}
	}
	return nil
}

func (s *FileStore) persist() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func validate(title, service string, severity Severity, status Status) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(service) == "" {
		return errors.New("service is required")
	}

	switch severity {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
	default:
		return errors.New("severity must be one of: low, medium, high, critical")
	}

	switch status {
	case StatusOpen, StatusInvestigating, StatusMitigated, StatusResolved:
	default:
		return errors.New("status must be one of: open, investigating, mitigated, resolved")
	}

	return nil
}

func matches(item Incident, options ListOptions) bool {
	if options.Service != "" && !strings.EqualFold(item.Service, options.Service) {
		return false
	}
	if options.Owner != "" && !strings.EqualFold(item.Owner, options.Owner) {
		return false
	}
	if options.Status != "" && item.Status != options.Status {
		return false
	}
	if options.Severity != "" && item.Severity != options.Severity {
		return false
	}
	if options.Search != "" {
		query := strings.ToLower(strings.TrimSpace(options.Search))
		haystack := strings.ToLower(strings.Join([]string{item.Title, item.Service, item.Description, item.Owner}, " "))
		if !strings.Contains(haystack, query) {
			return false
		}
	}
	return true
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func ParseListOptions(params map[string]string) ListOptions {
	options := ListOptions{
		Service:  strings.TrimSpace(params["service"]),
		Status:   Status(strings.TrimSpace(params["status"])),
		Severity: Severity(strings.TrimSpace(params["severity"])),
		Owner:    strings.TrimSpace(params["owner"]),
		Search:   strings.TrimSpace(params["search"]),
	}
	if limit, err := strconv.Atoi(strings.TrimSpace(params["limit"])); err == nil {
		options.Limit = limit
	}
	if offset, err := strconv.Atoi(strings.TrimSpace(params["offset"])); err == nil {
		options.Offset = offset
	}
	return options
}
