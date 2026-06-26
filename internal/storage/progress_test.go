package storage

import "testing"

func TestBuildProgressQueryArgs(t *testing.T) {
	args, err := buildProgressQueryArgs([]string{"userId", "graphCode", "student_id"}, "graph-a", "20240001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if args[0] != "20240001" || args[1] != "graph-a" || args[2] != "20240001" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildProgressQueryArgsRejectsUnknownToken(t *testing.T) {
	if _, err := buildProgressQueryArgs([]string{"classId"}, "graph-a", "20240001"); err == nil {
		t.Fatal("expected unsupported token error")
	}
}

func TestProgressCompletionQueryUsesDefaultAndCustomSQL(t *testing.T) {
	store := &Store{}
	query, args, source := store.progressCompletionQuery()
	if query == "" {
		t.Fatal("expected default query")
	}
	if source != "student_assignment" {
		t.Fatalf("expected default source, got %s", source)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 default args, got %d", len(args))
	}

	store.ConfigureProgressCompletionQuery("SELECT experiment_id FROM custom WHERE student_id = ?", []string{"userId"})
	query, args, source = store.progressCompletionQuery()
	if query != "SELECT experiment_id FROM custom WHERE student_id = ?" {
		t.Fatalf("unexpected custom query: %s", query)
	}
	if source != "custom_sql" {
		t.Fatalf("expected custom source, got %s", source)
	}
	if len(args) != 1 || args[0] != "userId" {
		t.Fatalf("unexpected custom args: %#v", args)
	}
}
