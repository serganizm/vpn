package store

import (
	"sync"
	"time"

	"github.com/serganizm/vpn/app/internal/models"
)

type Store struct {
	mu       sync.RWMutex
	pairs    map[string]*models.Pair
	users    map[int64]*models.User
	profiles []*models.Profile
	jobs     map[string]*models.InstallJob
	nextUID  int64
}

func New() *Store {
	s := &Store{
		pairs:    map[string]*models.Pair{},
		users:    map[int64]*models.User{},
		profiles: []*models.Profile{},
		jobs:     map[string]*models.InstallJob{},
		nextUID:  1,
	}
	return s
}

func (s *Store) ListPairs() []*models.Pair {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Pair, 0, len(s.pairs))
	for _, p := range s.pairs {
		cp := *p
		out = append(out, &cp)
	}
	return out
}

func (s *Store) GetPair(id string) (*models.Pair, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pairs[id]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

func (s *Store) UpsertPair(p *models.Pair) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	cp := *p
	s.pairs[p.ID] = &cp
}

func (s *Store) SetPairSession(id string, on bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pairs[id]
	if !ok {
		return false
	}
	p.Session = on
	p.UpdatedAt = time.Now().UTC()
	return true
}

func (s *Store) ListUsers(pairID string) []*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*models.User{}
	for _, u := range s.users {
		if pairID == "" || u.PairID == pairID {
			cp := *u
			out = append(out, &cp)
		}
	}
	return out
}

func (s *Store) GetUser(id int64) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (s *Store) AddUser(u *models.User) *models.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u.ID == 0 {
		u.ID = s.nextUID
		s.nextUID++
	}
	if u.Connections == nil {
		u.Connections = []models.Connection{}
	}
	cp := *u
	s.users[u.ID] = &cp
	return &cp
}

func (s *Store) UpdateUser(u *models.User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.ID]; !ok {
		return false
	}
	cp := *u
	if cp.Connections == nil {
		cp.Connections = []models.Connection{}
	}
	s.users[u.ID] = &cp
	return true
}

func (s *Store) DeleteUser(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return false
	}
	delete(s.users, id)
	return true
}

func (s *Store) ListProfiles() []*models.Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Profile, len(s.profiles))
	copy(out, s.profiles)
	return out
}

func (s *Store) AddProfile(p models.Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles = append(s.profiles, &models.Profile{ID: p.ID, Name: p.Name})
}

func (s *Store) PutJob(j *models.InstallJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *j
	s.jobs[j.ID] = &cp
}

func (s *Store) GetJob(id string) (*models.InstallJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}
