package practice

import (
	"strings"
	"testing"
)

func TestDescribeQuizResultGoodWithTopics(t *testing.T) {
	pct, explanation, _, review := DescribeQuizResult(8, 10, []string{"Deploy y release", "Bugs y fixes", "Deploy y release"})
	if pct != 80 {
		t.Fatalf("pct %d", pct)
	}
	if !strings.Contains(explanation, "está bien") || !strings.Contains(explanation, "Deploy y release") || !strings.Contains(explanation, "Bugs y fixes") {
		t.Fatalf("explanation %q", explanation)
	}
	if strings.Count(explanation, "Deploy y release") != 1 {
		t.Fatalf("topic duplicated: %q", explanation)
	}
	if !strings.Contains(review, "Deploy y release") || !strings.Contains(review, "Bugs y fixes") {
		t.Fatalf("review %q", review)
	}
}

func TestDescribeQuizResultPerfect(t *testing.T) {
	pct, explanation, _, review := DescribeQuizResult(10, 10, nil)
	if pct != 100 {
		t.Fatalf("pct %d", pct)
	}
	if !strings.Contains(explanation, "está bien") || strings.Contains(explanation, "Conviene reforzar:") {
		t.Fatalf("explanation %q", explanation)
	}
	if review != "No hay conceptos pendientes en esta sesión." {
		t.Fatalf("review %q", review)
	}
}

func TestDescribeQuizResultBelowThreshold(t *testing.T) {
	pct, explanation, _, _ := DescribeQuizResult(5, 10, []string{"APIs y servicios"})
	if pct != 50 {
		t.Fatalf("pct %d", pct)
	}
	if !strings.Contains(explanation, "bajo 60") || !strings.Contains(explanation, "APIs y servicios") {
		t.Fatalf("explanation %q", explanation)
	}
}
