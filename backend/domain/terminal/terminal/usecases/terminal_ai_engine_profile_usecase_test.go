package usecases

import (
	"errors"
	"sensio/domain/common/utils"
	"sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/entities"
	"testing"
)

// --- stubTerminalRepo for usecase tests ---

type stubAIProfileRepo struct {
	terminal *entities.Terminal
	notFound bool
	saveErr  error
}

func (r *stubAIProfileRepo) Create(t *entities.Terminal) error    { return nil }
func (r *stubAIProfileRepo) GetAll() ([]entities.Terminal, error) { return nil, nil }
func (r *stubAIProfileRepo) GetAllPaginated(o, l int, roomID *string) ([]entities.Terminal, int64, error) {
	return nil, 0, nil
}
func (r *stubAIProfileRepo) GetByMacAddress(mac string) (*entities.Terminal, error) {
	if r.notFound {
		return nil, errors.New("record not found")
	}
	return r.terminal, nil
}
func (r *stubAIProfileRepo) GetByID(id string) (*entities.Terminal, error) {
	if r.notFound {
		return nil, errors.New("record not found")
	}
	return r.terminal, nil
}
func (r *stubAIProfileRepo) GetByRoomID(roomID string) ([]entities.Terminal, error) { return nil, nil }
func (r *stubAIProfileRepo) Update(t *entities.Terminal) error                      { return r.saveErr }
func (r *stubAIProfileRepo) Delete(id string) error                                 { return nil }
func (r *stubAIProfileRepo) InvalidateCache(id string) error                        { return nil }
func (r *stubAIProfileRepo) CreateMQTTUser(u *entities.MQTTUser) error              { return nil }
func (r *stubAIProfileRepo) GetMQTTUserByUsername(username string) (*entities.MQTTUser, error) {
	return nil, nil
}

func strPtr(s string) *string { return &s }

var validTestConfig = &utils.Config{
	OpenAIApiKey: "dummy",
	GroqApiKey:   "dummy",
	OrionApiKey:  "dummy",
}

// --- Update use case tests ---

func TestUpdateAIEngineProfile_AcceptsFast(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1"}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	result, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("fast")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile == nil || *result.Profile != "fast" {
		t.Errorf("expected profile=fast, got %v", result.Profile)
	}
}

func TestUpdateAIEngineProfile_AcceptsStandard(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1"}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	result, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("standard")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile == nil || *result.Profile != "standard" {
		t.Errorf("expected profile=standard, got %v", result.Profile)
	}
}

func TestUpdateAIEngineProfile_RejectsPlaud(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1"}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	_, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("plaud")})
	if err == nil {
		t.Fatal("expected validation error for plaud")
	}
}

func TestUpdateAIEngineProfile_RejectsUnknownValue(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1"}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	_, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("turbo")})
	if err == nil {
		t.Fatal("expected validation error for unknown profile")
	}
}

func TestUpdateAIEngineProfile_ClearsOnEmpty(t *testing.T) {
	existing := "fast"
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1", AiEngineProfile: &existing}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	empty := ""
	result, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: &empty})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile != nil {
		t.Errorf("expected profile=nil after clear, got %v", *result.Profile)
	}
}

func TestUpdateAIEngineProfile_ClearsOnNull(t *testing.T) {
	existing := "standard"
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1", AiEngineProfile: &existing}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	result, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile != nil {
		t.Errorf("expected profile=nil after null clear, got %v", *result.Profile)
	}
}

func TestUpdateAIEngineProfile_NotFound(t *testing.T) {
	repo := &stubAIProfileRepo{notFound: true}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, validTestConfig)

	_, err := uc.Update("no-such-id", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("fast")})
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if err.Error() != "Terminal not found" {
		t.Errorf("expected 'Terminal not found', got %q", err.Error())
	}
}

func TestUpdateAIEngineProfile_RejectsFastWhenNotConfigured(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1"}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, &utils.Config{}) // empty config

	_, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("fast")})
	if err == nil {
		t.Fatal("expected validation error for fast without config")
	}
}

func TestUpdateAIEngineProfile_RejectsStandardWhenNotConfigured(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{ID: "t1"}}
	uc := NewUpdateTerminalAIEngineProfileUseCase(repo, &utils.Config{})

	_, err := uc.Update("t1", &dtos.UpdateTerminalAIEngineProfileRequestDTO{Profile: strPtr("standard")})
	if err == nil {
		t.Fatal("expected validation error for standard without config")
	}
}

// --- Get use case tests ---

func TestGetAIEngineProfile_ReturnsProfile(t *testing.T) {
	profile := "fast"
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{
		ID:              "t1",
		AiEngineProfile: &profile,
		MacAddress:      "AA:BB:CC",
	}}
	uc := NewGetTerminalAIEngineProfileUseCase(repo)

	result, err := uc.GetByMac("AA:BB:CC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile == nil || *result.Profile != "fast" {
		t.Errorf("expected profile=fast, got %v", result.Profile)
	}
	if result.Source != "engine_profile" {
		t.Errorf("expected source=engine_profile, got %s", result.Source)
	}
	if result.EffectiveMode != "fast" {
		t.Errorf("expected effective_mode=fast, got %s", result.EffectiveMode)
	}
}

func TestGetAIEngineProfile_LegacyProviderFallback(t *testing.T) {
	provider := "openai"
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{
		ID:         "t1",
		AiProvider: &provider,
		MacAddress: "AA:BB:CC",
	}}
	uc := NewGetTerminalAIEngineProfileUseCase(repo)

	result, err := uc.GetByMac("AA:BB:CC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile != nil {
		t.Errorf("expected profile=nil, got %v", *result.Profile)
	}
	if result.Source != "legacy_provider" {
		t.Errorf("expected source=legacy_provider, got %s", result.Source)
	}
	if result.EffectiveProvider == nil || *result.EffectiveProvider != "openai" {
		t.Errorf("expected effective_provider=openai, got %v", result.EffectiveProvider)
	}
}

func TestGetAIEngineProfile_DefaultPath(t *testing.T) {
	repo := &stubAIProfileRepo{terminal: &entities.Terminal{
		ID:         "t1",
		MacAddress: "AA:BB:CC",
	}}
	uc := NewGetTerminalAIEngineProfileUseCase(repo)

	result, err := uc.GetByMac("AA:BB:CC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != "default" {
		t.Errorf("expected source=default, got %s", result.Source)
	}
}

func TestGetAIEngineProfile_NotFound(t *testing.T) {
	repo := &stubAIProfileRepo{notFound: true}
	uc := NewGetTerminalAIEngineProfileUseCase(repo)

	_, err := uc.GetByMac("no-mac")
	if err == nil {
		t.Fatal("expected not-found error")
	}
}
