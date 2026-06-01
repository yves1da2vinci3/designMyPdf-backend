package usercredit

import (
	"fmt"
	"time"
)

// ConsumeRequest carries per-model token usage for precise credit deduction.
type ConsumeRequest struct {
	Model        string
	InputTokens  int
	OutputTokens int
}

// modelRates maps model ID prefixes to [inputRate, outputRate] in µcredits/token.
// µcredits/token == dollar price per 1M tokens (numerically identical).
// 1 credit = 1000 µcredits = $0.001. $1 limit = 1,000,000 µcredits.
var modelRates = map[string][2]int{
	"claude-haiku-4-5-20251001": {1, 5},
	"claude-haiku-4-5":          {1, 5},
	"claude-sonnet-4-20250514":  {3, 15},
	"claude-sonnet-4-5":         {3, 15},
	"claude-sonnet-4-6":         {3, 15},
	"claude-opus-4-7":           {5, 25},
	"claude-opus-4-8":           {5, 25},
}

// defaultRate is used when model is unknown — Sonnet pricing as safe fallback.
var defaultRate = [2]int{3, 15}

func calcMicroCredits(model string, inputTokens, outputTokens int) int {
	rate, ok := modelRates[model]
	if !ok {
		rate = defaultRate
	}
	return inputTokens*rate[0] + outputTokens*rate[1]
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func currentMonth() string {
	return time.Now().UTC().Format("2006-01")
}

// GetBalance returns µcredits used/limit/remaining and float display values (in credits).
func (s *Service) GetBalance(userID uint) (used, limit, remaining int, creditsUsed, creditsLimit, creditsRemaining float64, month string, err error) {
	month = currentMonth()
	uc, err := s.repo.GetOrCreate(userID, month)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, month, err
	}
	used = uc.CreditsUsed
	limit = uc.CreditsLimit
	remaining = limit - used
	if remaining < 0 {
		remaining = 0
	}
	creditsUsed = float64(used) / 1000
	creditsLimit = float64(limit) / 1000
	creditsRemaining = float64(remaining) / 1000
	return used, limit, remaining, creditsUsed, creditsLimit, creditsRemaining, month, nil
}

// ConsumeResult is returned by Consume / ConsumeUpToLimit.
type ConsumeResult struct {
	RemainingMicro   int
	CreditsRemaining float64
	DeductedMicro    int
	Capped           bool
}

// Consume deducts µcredits based on model pricing. Returns (µcreditsRemaining, creditsRemaining, error).
func (s *Service) Consume(userID uint, req ConsumeRequest) (int, float64, error) {
	result, err := s.consume(userID, req, false)
	if err != nil {
		return 0, 0, err
	}
	return result.RemainingMicro, result.CreditsRemaining, nil
}

// ConsumeUpToLimit deducts at most the remaining monthly budget (no error when capped).
func (s *Service) ConsumeUpToLimit(userID uint, req ConsumeRequest) (ConsumeResult, error) {
	return s.consume(userID, req, true)
}

func (s *Service) consume(userID uint, req ConsumeRequest, allowPartial bool) (ConsumeResult, error) {
	month := currentMonth()
	uc, err := s.repo.GetOrCreate(userID, month)
	if err != nil {
		return ConsumeResult{}, err
	}

	toDeduct := calcMicroCredits(req.Model, req.InputTokens, req.OutputTokens)
	available := uc.CreditsLimit - uc.CreditsUsed
	if available < 0 {
		available = 0
	}

	if toDeduct <= 0 {
		return ConsumeResult{
			RemainingMicro:   available,
			CreditsRemaining: float64(available) / 1000,
			DeductedMicro:    0,
			Capped:           false,
		}, nil
	}

	actual := toDeduct
	capped := false
	if uc.CreditsUsed+toDeduct > uc.CreditsLimit {
		if !allowPartial {
			return ConsumeResult{
				RemainingMicro:   available,
				CreditsRemaining: float64(available) / 1000,
			}, fmt.Errorf("monthly credit limit reached")
		}
		actual = available
		capped = actual < toDeduct
	}

	uc.CreditsUsed += actual
	if err := s.repo.Save(uc); err != nil {
		return ConsumeResult{}, err
	}

	remaining := uc.CreditsLimit - uc.CreditsUsed
	if remaining < 0 {
		remaining = 0
	}
	return ConsumeResult{
		RemainingMicro:   remaining,
		CreditsRemaining: float64(remaining) / 1000,
		DeductedMicro:    actual,
		Capped:           capped,
	}, nil
}
