package practice

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"
)

const (
	scoreThreshold    = 60
	recurrenceWindow  = 5
	recurrenceMinLow  = 3
	statusProvisional = "provisional"
	statusStable      = "stable"
	statusRecurrent   = "recurrent"
)

// ProgressSnapshot es una practica ya evaluada, de la mas reciente a la mas antigua.
// Un puntero nil significa que esa practica no evaluo la competencia.
type ProgressSnapshot struct {
	Comprehension *int
	Vocabulary    *int
	Communication *int
}

// CompetencyAverage resume el historial de una competencia.
type CompetencyAverage struct {
	Competency string `json:"competency"`
	Label      string `json:"label"`
	Average    *int   `json:"average"`
	ValidCount int    `json:"validCount"`
}

// Recommendation es el resultado de las reglas propias de refuerzo.
type Recommendation struct {
	Status          string              `json:"status"`
	Message         string              `json:"message"`
	Competency      string              `json:"competency,omitempty"`
	CompetencyLabel string              `json:"competencyLabel,omitempty"`
	CategoryCode    string              `json:"categoryCode,omitempty"`
	CategoryName    string              `json:"categoryName,omitempty"`
	Averages        []CompetencyAverage `json:"averages"`
}

type competencyRule struct {
	key        string
	label      string
	module     string
	moduleName string
	pick       func(ProgressSnapshot) *int
}

func competencyRules() []competencyRule {
	return []competencyRule{
		{
			key: "comprehension", label: "Comprensión",
			module: "daily_meeting", moduleName: "Daily meeting",
			pick: func(s ProgressSnapshot) *int { return s.Comprehension },
		},
		{
			key: "vocabulary", label: "Vocabulario profesional",
			module: "it_quiz", moduleName: "Quiz de inglés IT",
			pick: func(s ProgressSnapshot) *int { return s.Vocabulary },
		},
		{
			key: "communication", label: "Comunicación profesional",
			module: "hr_interview", moduleName: "Entrevista con Recursos Humanos",
			pick: func(s ProgressSnapshot) *int { return s.Communication },
		},
	}
}

type recurrenceCandidate struct {
	index int
	avg   float64
	low   int
}

// EvaluateProgress aplica la regla del prototipo:
// dificultad recurrente si una competencia queda bajo 60 en al menos tres
// de las ultimas cinco actividades que la evaluan. Con menos de cinco
// resultados validos esa competencia no se clasifica.
// Si hay varias, gana el menor promedio de la ventana; luego la que tenga
// mas resultados insuficientes; luego el orden comprension, vocabulario, comunicacion.
func EvaluateProgress(newestFirst []ProgressSnapshot) Recommendation {
	rules := competencyRules()
	averages := make([]CompetencyAverage, 0, len(rules))
	var candidates []recurrenceCandidate
	evaluated := 0
	complete := 0

	for i, rule := range rules {
		var values []int
		for _, snap := range newestFirst {
			if v := rule.pick(snap); v != nil {
				values = append(values, *v)
			}
		}

		item := CompetencyAverage{
			Competency: rule.key,
			Label:      rule.label,
			ValidCount: len(values),
		}
		if len(values) > 0 {
			evaluated++
			sum := 0
			for _, n := range values {
				sum += n
			}
			rounded := int(math.Round(float64(sum) / float64(len(values))))
			item.Average = &rounded
		}
		averages = append(averages, item)

		window := values
		if len(window) > recurrenceWindow {
			window = window[:recurrenceWindow]
		}
		if len(window) < recurrenceWindow {
			continue
		}
		complete++
		low := 0
		sum := 0
		for _, n := range window {
			sum += n
			if n < scoreThreshold {
				low++
			}
		}
		if low >= recurrenceMinLow {
			candidates = append(candidates, recurrenceCandidate{
				index: i,
				avg:   float64(sum) / float64(len(window)),
				low:   low,
			})
		}
	}

	rec := Recommendation{Averages: averages}
	if len(candidates) > 0 {
		sort.SliceStable(candidates, func(i, j int) bool {
			a, b := candidates[i], candidates[j]
			if a.avg != b.avg {
				return a.avg < b.avg
			}
			if a.low != b.low {
				return a.low > b.low
			}
			return a.index < b.index
		})
		best := rules[candidates[0].index]
		rec.Status = statusRecurrent
		rec.Competency = best.key
		rec.CompetencyLabel = best.label
		rec.CategoryCode = best.module
		rec.CategoryName = best.moduleName
		rec.Message = fmt.Sprintf(
			"%s quedó bajo 60 en al menos tres de las últimas cinco prácticas. Te recomendamos reforzar con %s.",
			best.label, best.moduleName,
		)
		return rec
	}

	if evaluated == 0 || complete < evaluated {
		rec.Status = statusProvisional
		rec.Message = "Con menos de cinco resultados válidos por competencia el progreso es provisorio. Seguí practicando: la recomendación de refuerzo aparece cuando una competencia queda bajo 60 en tres de las últimas cinco."
		return rec
	}

	rec.Status = statusStable
	rec.Message = "Ninguna competencia quedó bajo 60 en tres de las últimas cinco prácticas. Seguí entrenando entrevista, daily meeting y quiz."
	return rec
}

// Recommend carga el historial del usuario y aplica EvaluateProgress.
func (s *Service) Recommend(ctx context.Context, userID uuid.UUID) (Recommendation, error) {
	items, err := s.repo.ListSessionsByUser(ctx, userID, maxListLimit, 0)
	if err != nil {
		return Recommendation{}, err
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	snaps := make([]ProgressSnapshot, 0, len(items))
	for _, item := range items {
		if item.Scores == nil {
			continue
		}
		snaps = append(snaps, ProgressSnapshot{
			Comprehension: item.Scores.Comprehension,
			Vocabulary:    item.Scores.Vocabulary,
			Communication: item.Scores.Communication,
		})
	}
	return EvaluateProgress(snaps), nil
}
