package organization

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/modura-dev/modura/backend/internal/modules/audit"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
)

type storeStub struct {
	department Department
	position   Position
}

func (*storeStub) ListDepartments(context.Context, identity.TenantID, DataScope) ([]DepartmentView, error) {
	return nil, nil
}
func (*storeStub) ListPositions(context.Context, identity.TenantID) ([]PositionView, error) {
	return nil, nil
}

func (s *storeStub) CreateDepartment(_ context.Context, _ pgx.Tx, department Department, _ DataScope) error {
	s.department = department
	return nil
}
func (*storeStub) MoveDepartment(context.Context, pgx.Tx, identity.TenantID, DepartmentID, DepartmentID, DataScope, time.Time) error {
	return nil
}
func (*storeStub) DeleteDepartment(context.Context, pgx.Tx, identity.TenantID, DepartmentID, DataScope) error {
	return nil
}
func (s *storeStub) CreatePosition(_ context.Context, _ pgx.Tx, position Position) error {
	s.position = position
	return nil
}
func (*storeStub) AssignUser(context.Context, pgx.Tx, identity.TenantID, identity.UserID, DepartmentID, *PositionID, DataScope, time.Time) error {
	return nil
}

type transactorStub struct{}

func (transactorStub) WithinTransaction(_ context.Context, work func(pgx.Tx) error) error {
	return work(nil)
}

type auditorStub struct{ events []audit.Event }

func (s *auditorStub) RecordTenantWrite(_ context.Context, _ pgx.Tx, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}
func (*storeStub) ProvisionInitialOrganization(context.Context, pgx.Tx, Department, identity.UserID) error {
	return nil
}

func TestServiceNormalizesOrganizationNames(t *testing.T) {
	store := &storeStub{}
	auditor := &auditorStub{}
	now := time.Unix(1_700_000_000, 0)
	service, err := NewService(store, transactorStub{}, auditor, func() time.Time { return now }, func(time.Time) (string, error) { return "018bcfe5-6800-7000-8000-000000000201", nil })
	if err != nil {
		t.Fatal(err)
	}
	write := WriteContext{Actor: identity.Actor{TenantID: "tenant", UserID: "user", SessionID: "session"}, CorrelationID: "request-1", Scope: DataScope{ActorID: "user", All: true}}
	if _, err := service.CreateDepartment(context.Background(), write, nil, " 研发中心 ", 10); err != nil {
		t.Fatal(err)
	}
	if store.department.Name != "研发中心" || store.department.NormalizedName != "研发中心" || !store.department.CreatedAt.Equal(now) {
		t.Fatalf("department = %+v", store.department)
	}
	if _, err := service.CreatePosition(context.Background(), write, " Platform ENGINEER "); err != nil {
		t.Fatal(err)
	}
	if store.position.Name != "Platform ENGINEER" || store.position.NormalizedName != "platform engineer" {
		t.Fatalf("position = %+v", store.position)
	}
	if len(auditor.events) != 2 || auditor.events[0].ActorID != "user" || auditor.events[0].CorrelationID != "request-1" {
		t.Fatalf("audit events = %+v", auditor.events)
	}
}
