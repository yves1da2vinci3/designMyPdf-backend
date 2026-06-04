// Package usercredit déduit un budget mensuel par utilisateur à partir des tokens API.
//
// Économie (miroir tarifs Anthropic $/million de tokens) :
//   - 1 crédit affiché = 1000 µcrédits = 0,001 $ de coût catalogue.
//   - Plafond mensuel = 1_000_000 µcrédits = 1000 crédits affichés = 1 $ max/utilisateur/mois.
//
// Débit : µcrédits = inputTokens×rateIn + outputTokens×rateOut
//   rateIn/rateOut = prix $ par 1M tokens (ex. Haiku {1,5}, Sonnet {3,15}).
//
// Exemple Sonnet (image) : 15k in + 6k out → 135_000 µcrédits = 135 crédits (~13,5 % du mois).
// Exemple Haiku (texte) : 8k in + 2k out → 18_000 µcrédits = 18 crédits.
package usercredit

import (
	"fmt"
	"strings"
	"time"
)

// ConsumeRequest carries per-model token usage for precise credit deduction.
type ConsumeRequest struct {
	Model        string
	InputTokens  int
	OutputTokens int
}

// modelRates : préfixe modèle → [tarif input, tarif output] en µcrédits/token (= $/1M tokens).
var modelRates = map[string][2]int{
	"claude-haiku-4-5-20251001": {1, 5},
	"claude-haiku-4-5":          {1, 5},
	"claude-sonnet-4-20250514":  {3, 15},
	"claude-sonnet-4-5":         {3, 15},
	"claude-sonnet-4-6":         {3, 15},
	"claude-sonnet-4":           {3, 15},
	"claude-3-5-sonnet-latest":  {3, 15},
	"claude-opus-4-7":           {5, 25},
	"claude-opus-4-8":           {5, 25},
}

// defaultRate is used when model is unknown — Sonnet pricing as safe fallback.
var defaultRate = [2]int{3, 15}

func resolveModelRate(model string) [2]int {
	bestLen := 0
	var best [2]int
	for prefix, rate := range modelRates {
		if strings.HasPrefix(model, prefix) && len(prefix) > bestLen {
			bestLen = len(prefix)
			best = rate
		}
	}
	if bestLen > 0 {
		return best
	}
	return defaultRate
}

func calcMicroCredits(model string, inputTokens, outputTokens int) int {
	rate := resolveModelRate(model)
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

// ConsumeWithResult deducts credits and returns full result including deducted amount.
func (s *Service) ConsumeWithResult(userID uint, req ConsumeRequest) (ConsumeResult, error) {
	return s.consume(userID, req, false)
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
