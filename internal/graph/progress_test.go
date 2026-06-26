package graph

import "testing"

func TestComputeProgressFromCompletedExercises(t *testing.T) {
	chapterLinear := "ds-ch-linear"
	chapterTree := "ds-ch-tree"
	g := &Graph{
		Course: &Node{ID: "ds-course", Label: "数据结构", Type: "course"},
		Nodes: []*Node{
			{ID: "ds-course", Label: "数据结构", Type: "course"},
			{ID: chapterLinear, Label: "线性表", Type: "chapter"},
			{ID: chapterTree, Label: "树", Type: "chapter"},
			{ID: "ds-struct-seqlist", Label: "顺序表", Type: "structure", ChapterID: &chapterLinear},
			{ID: "ds-struct-tree", Label: "树结构", Type: "structure", ChapterID: &chapterTree},
			{
				ID:         "ds-ex-1",
				Label:      "第1次作业",
				Type:       "exercise",
				ChapterID:  &chapterLinear,
				Properties: MapRaw{"experimentId": float64(1)},
				Targets:    []string{"ds-struct-seqlist"},
			},
			{
				ID:         "ds-ex-2",
				Label:      "第1次实验",
				Type:       "exercise",
				ChapterID:  &chapterLinear,
				Properties: MapRaw{"experimentId": float64(2)},
			},
			{
				ID:         "ds-ex-3",
				Label:      "第2次实验",
				Type:       "exercise",
				ChapterID:  &chapterTree,
				Properties: MapRaw{"experimentId": float64(3)},
				Targets:    []string{"ds-struct-tree"},
			},
		},
		Relations: []*Relation{
			{ID: "r1", Source: "ds-struct-seqlist", Target: "ds-ex-2", Type: "TESTED_BY"},
			{ID: "r1-dup", Source: "ds-struct-seqlist", Target: "ds-ex-2", Type: "TESTED_BY"},
		},
	}

	progress := ComputeProgress("data-structure-knowledge-graph", g, "20240001", []string{"1", "ds-ex-2"}, "test")

	assertNodeProgress(t, progress.Nodes["ds-ex-1"], 100, LearningStatusMastered, 1, 1)
	assertNodeProgress(t, progress.Nodes["ds-ex-2"], 100, LearningStatusMastered, 1, 1)
	assertNodeProgress(t, progress.Nodes["ds-ex-3"], 0, LearningStatusUnstarted, 0, 1)
	assertNodeProgress(t, progress.Nodes["ds-struct-seqlist"], 100, LearningStatusMastered, 2, 2)
	assertNodeProgress(t, progress.Nodes["ds-struct-tree"], 0, LearningStatusUnstarted, 0, 1)
	assertNodeProgress(t, progress.Nodes[chapterLinear], 100, LearningStatusMastered, 2, 2)
	assertNodeProgress(t, progress.Nodes[chapterTree], 0, LearningStatusUnstarted, 0, 1)
	assertNodeProgress(t, progress.Nodes["ds-course"], 66, LearningStatusLearning, 2, 3)

	if progress.ProgressPercent != 66 {
		t.Fatalf("expected graph percent 66, got %d", progress.ProgressPercent)
	}
	if progress.LearningStatus != LearningStatusLearning {
		t.Fatalf("expected graph status %s, got %s", LearningStatusLearning, progress.LearningStatus)
	}
}

func TestApplyProgressToGraph(t *testing.T) {
	chapterID := "ds-ch-linear"
	g := &Graph{
		Course: &Node{ID: "ds-course", Label: "数据结构", Type: "course"},
		Nodes: []*Node{
			{ID: chapterID, Label: "线性表", Type: "chapter"},
			{
				ID:         "ds-ex-1",
				Label:      "第1次作业",
				Type:       "exercise",
				ChapterID:  &chapterID,
				Properties: MapRaw{"experimentId": 1},
			},
		},
	}
	progress := ComputeProgress("data-structure-knowledge-graph", g, "20240001", []string{"experiment:1"}, "test")

	ApplyProgressToGraph(g, progress)

	if g.Course.ProgressPercent == nil || *g.Course.ProgressPercent != 100 {
		t.Fatalf("expected course progress 100, got %#v", g.Course.ProgressPercent)
	}
	exercise := g.Nodes[1]
	if exercise.LearningStatus != LearningStatusMastered {
		t.Fatalf("expected exercise status %s, got %s", LearningStatusMastered, exercise.LearningStatus)
	}
	if exercise.CompletedExerciseCount == nil || *exercise.CompletedExerciseCount != 1 {
		t.Fatalf("expected exercise completed count 1, got %#v", exercise.CompletedExerciseCount)
	}
}

func TestComputeProgressWithNoExercises(t *testing.T) {
	g := &Graph{
		Course: &Node{ID: "ds-course", Label: "数据结构", Type: "course"},
		Nodes: []*Node{
			{ID: "ds-ch-linear", Label: "线性表", Type: "chapter"},
			{ID: "ds-concept-complexity", Label: "复杂度", Type: "concept"},
		},
	}

	progress := ComputeProgress("data-structure-knowledge-graph", g, "20240001", nil, "test")

	assertNodeProgress(t, progress.Nodes["ds-course"], 0, LearningStatusUnstarted, 0, 0)
	assertNodeProgress(t, progress.Nodes["ds-ch-linear"], 0, LearningStatusUnstarted, 0, 0)
	assertNodeProgress(t, progress.Nodes["ds-concept-complexity"], 0, LearningStatusUnstarted, 0, 0)
}

func assertNodeProgress(t *testing.T, progress *NodeProgress, percent int, status string, completed int, total int) {
	t.Helper()
	if progress == nil {
		t.Fatal("expected node progress, got nil")
	}
	if progress.ProgressPercent != percent {
		t.Fatalf("expected percent %d, got %d for node %s", percent, progress.ProgressPercent, progress.NodeID)
	}
	if progress.LearningStatus != status {
		t.Fatalf("expected status %s, got %s for node %s", status, progress.LearningStatus, progress.NodeID)
	}
	if progress.CompletedExerciseCount != completed {
		t.Fatalf("expected completed %d, got %d for node %s", completed, progress.CompletedExerciseCount, progress.NodeID)
	}
	if progress.TotalExerciseCount != total {
		t.Fatalf("expected total %d, got %d for node %s", total, progress.TotalExerciseCount, progress.NodeID)
	}
}
