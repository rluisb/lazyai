package agentmemory

import "testing"

func TestOpenRunsMigrationsIdempotently(t *testing.T) {
	db := testDB(t)

	for _, table := range []string{"tasks", "task_events", "checkpoints", "artifacts", "memories", "memories_fts", "artifacts_fts", "_migrations"} {
		var name string
		if err := db.QueryRow("SELECT name FROM sqlite_master WHERE name = ?", table).Scan(&name); err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations() second run error = %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != len(migrations) {
		t.Fatalf("migration count = %d, want %d", count, len(migrations))
	}
}
