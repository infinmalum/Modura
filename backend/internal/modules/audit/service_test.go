package audit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

type recordStoreStub struct{}

func (recordStoreStub) Record(context.Context, pgx.Tx, Event) error                 { return nil }
func (recordStoreStub) RecordPlatform(context.Context, pgx.Tx, PlatformEvent) error { return nil }

type queryStoreStub struct{ records []Record }

func (s queryStoreStub) List(context.Context, Query) ([]Record, error) { return s.records, nil }

func TestListRedactsSensitiveStateRecursively(t *testing.T) {
	service, err := NewService(recordStoreStub{}, func(time.Time) (string, error) { return "id", nil })
	if err != nil {
		t.Fatal(err)
	}
	if err := service.EnableQueries(queryStoreStub{records: []Record{{BeforeState: []byte(`{"profile":{"name":"Ada","refreshToken":"secret"},"session_id":"hidden"}`)}}}); err != nil {
		t.Fatal(err)
	}
	records, err := service.List(context.Background(), identity.TenantID("tenant"), "", "", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	state := string(records[0].BeforeState)
	if strings.Contains(state, "secret") || strings.Contains(state, "hidden") || !strings.Contains(state, "Ada") || strings.Count(state, "[redacted]") != 2 {
		t.Fatalf("redacted state = %s", state)
	}
}
