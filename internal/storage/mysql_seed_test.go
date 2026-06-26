package storage

import (
	"strings"
	"testing"

	seeddata "knowledgegraph/deploy/mysql/init"
)

func TestEmbeddedDefaultGraphSeedScript(t *testing.T) {
	script := seeddata.DataStructureGraphSQL
	if !strings.Contains(script, defaultGraphCode) {
		t.Fatalf("seed script does not contain default graph code %q", defaultGraphCode)
	}

	statements := splitSQLStatements(stripSQLLineComments(script))
	if len(statements) == 0 {
		t.Fatal("seed script produced no executable statements")
	}
	for _, statement := range statements {
		if strings.HasPrefix(strings.TrimSpace(statement), "--") {
			t.Fatalf("seed statement still starts with SQL comment: %q", statement)
		}
	}

	if count := strings.Count(script, "__TESTED_BY__ds-ex-"); count != 40 {
		t.Fatalf("expected 40 exercise relations, got %d", count)
	}
}
