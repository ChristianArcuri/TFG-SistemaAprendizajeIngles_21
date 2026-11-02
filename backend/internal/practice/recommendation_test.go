package practice

import "testing"

func ptr(n int) *int { return &n }

func snap(comp, voc, comm *int) ProgressSnapshot {
	return ProgressSnapshot{Comprehension: comp, Vocabulary: voc, Communication: comm}
}

func TestEvaluateProgressProvisionalBelowFive(t *testing.T) {
	var snaps []ProgressSnapshot
	for i := 0; i < 4; i++ {
		snaps = append(snaps, snap(ptr(40), ptr(40), ptr(40)))
	}
	got := EvaluateProgress(snaps)
	if got.Status != statusProvisional {
		t.Fatalf("status %s", got.Status)
	}
	if got.CategoryCode != "" {
		t.Fatalf("did not expect a module, got %s", got.CategoryCode)
	}
}

func TestEvaluateProgressStableWithTwoLows(t *testing.T) {
	snaps := []ProgressSnapshot{
		snap(ptr(40), ptr(80), ptr(80)),
		snap(ptr(40), ptr(80), ptr(80)),
		snap(ptr(80), ptr(80), ptr(80)),
		snap(ptr(80), ptr(80), ptr(80)),
		snap(ptr(80), ptr(80), ptr(80)),
	}
	got := EvaluateProgress(snaps)
	if got.Status != statusStable {
		t.Fatalf("status %s, message %s", got.Status, got.Message)
	}
}

func TestEvaluateProgressRecurrentAtThreeLows(t *testing.T) {
	snaps := []ProgressSnapshot{
		snap(ptr(80), ptr(40), ptr(90)),
		snap(ptr(80), ptr(40), ptr(90)),
		snap(ptr(80), ptr(40), ptr(90)),
		snap(ptr(80), ptr(80), ptr(90)),
		snap(ptr(80), ptr(80), ptr(90)),
	}
	got := EvaluateProgress(snaps)
	if got.Status != statusRecurrent || got.Competency != "vocabulary" || got.CategoryCode != "it_quiz" {
		t.Fatalf("got %+v", got)
	}
}

func TestEvaluateProgressTieUsesOrder(t *testing.T) {
	snaps := []ProgressSnapshot{
		snap(ptr(40), ptr(40), ptr(40)),
		snap(ptr(40), ptr(40), ptr(40)),
		snap(ptr(40), ptr(40), ptr(40)),
		snap(ptr(80), ptr(80), ptr(80)),
		snap(ptr(80), ptr(80), ptr(80)),
	}
	got := EvaluateProgress(snaps)
	if got.Competency != "comprehension" || got.CategoryCode != "daily_meeting" {
		t.Fatalf("expected comprehension first, got %+v", got)
	}
}

func TestEvaluateProgressLowerAverageWins(t *testing.T) {
	snaps := []ProgressSnapshot{
		snap(ptr(50), ptr(20), ptr(90)),
		snap(ptr(50), ptr(20), ptr(90)),
		snap(ptr(50), ptr(20), ptr(90)),
		snap(ptr(90), ptr(90), ptr(90)),
		snap(ptr(90), ptr(90), ptr(90)),
	}
	got := EvaluateProgress(snaps)
	if got.Competency != "vocabulary" {
		t.Fatalf("expected vocabulary, got %s", got.Competency)
	}
}

func TestEvaluateProgressQuizDoesNotCountCommunication(t *testing.T) {
	var snaps []ProgressSnapshot
	for i := 0; i < 5; i++ {
		snaps = append(snaps, snap(ptr(30), ptr(30), nil))
	}
	got := EvaluateProgress(snaps)
	if got.Status != statusRecurrent {
		t.Fatalf("status %s", got.Status)
	}
	if got.Competency != "comprehension" {
		t.Fatalf("expected comprehension over vocabulary on tie, got %s", got.Competency)
	}
	for _, avg := range got.Averages {
		if avg.Competency == "communication" && avg.ValidCount != 0 {
			t.Fatalf("communication counted quiz rows: %+v", avg)
		}
	}
}

func TestEvaluateProgressNewResultChangesWindow(t *testing.T) {
	base := []ProgressSnapshot{
		snap(ptr(40), ptr(90), ptr(90)),
		snap(ptr(40), ptr(90), ptr(90)),
		snap(ptr(40), ptr(90), ptr(90)),
		snap(ptr(90), ptr(90), ptr(90)),
		snap(ptr(90), ptr(90), ptr(90)),
	}
	if got := EvaluateProgress(base); got.Competency != "comprehension" {
		t.Fatalf("setup: %+v", got)
	}

	// Tres resultados altos nuevos dejan solo dos bajos dentro de la ventana.
	updated := []ProgressSnapshot{
		snap(ptr(95), ptr(90), ptr(90)),
		snap(ptr(95), ptr(90), ptr(90)),
		snap(ptr(95), ptr(90), ptr(90)),
		snap(ptr(40), ptr(90), ptr(90)),
		snap(ptr(40), ptr(90), ptr(90)),
	}
	got := EvaluateProgress(updated)
	if got.Status != statusStable {
		t.Fatalf("expected window to drop the third low, got %+v", got)
	}
}
