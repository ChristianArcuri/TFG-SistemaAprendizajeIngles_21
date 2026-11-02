package practice

import (
	"fmt"
	"strings"
)

// DescribeQuizResult arma el puntaje y el texto del quiz sin llamar a la IA.
// 80 o más se considera un buen resultado. Si hubo errores, igual se nombran
// los temas para reforzar. Menos de 60 queda como resultado insuficiente.
func DescribeQuizResult(correct, total int, missed []string) (int, string, string, string) {
	pct := 0
	if total > 0 {
		pct = clampScore(int(float64(correct)*100/float64(total) + 0.5))
	}
	topics := uniqueTopics(missed)

	var explanation string
	switch {
	case correct == total:
		explanation = fmt.Sprintf("Respondiste %d de %d. El resultado está bien y no quedaron conceptos para reforzar en esta sesión.", correct, total)
	case pct >= 80 && len(topics) > 0:
		explanation = fmt.Sprintf("Respondiste %d de %d. Dentro de todo está bien. Conviene reforzar: %s.", correct, total, strings.Join(topics, ", "))
	case pct >= 60 && len(topics) > 0:
		explanation = fmt.Sprintf("Respondiste %d de %d. El resultado es suficiente. Conviene reforzar: %s.", correct, total, strings.Join(topics, ", "))
	case len(topics) > 0:
		explanation = fmt.Sprintf("Respondiste %d de %d. Quedó bajo 60. Conviene reforzar: %s.", correct, total, strings.Join(topics, ", "))
	default:
		explanation = fmt.Sprintf("Respondiste %d de %d. Conviene repasar las preguntas que no salieron bien.", correct, total)
	}

	summary := fmt.Sprintf("%d de %d respuestas correctas.", correct, total)
	review := "No hay conceptos pendientes en esta sesión."
	if len(topics) > 0 {
		review = "Conviene repasar: " + strings.Join(topics, ", ") + "."
	}
	return pct, explanation, summary, review
}

func uniqueTopics(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		topic := strings.TrimSpace(raw)
		if topic == "" || seen[topic] {
			continue
		}
		if len(topic) > 80 {
			topic = topic[:80]
		}
		seen[topic] = true
		out = append(out, topic)
		if len(out) == 8 {
			break
		}
	}
	return out
}
