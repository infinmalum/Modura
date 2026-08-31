package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	"github.com/modura-dev/modura/backend/internal/modules/organization"
)

func TestOrganizationTenantAndTreeInvariants(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0).UTC()
	seedIdentity(t, pool, now)
	store := New(pool)
	alpha := identity.TenantID("018bcfe5-6800-7000-8000-000000000101")
	beta := identity.TenantID("018bcfe5-6800-7000-8000-000000000102")
	alphaRoot := department("018bcfe5-6800-7000-8000-000000000111", alpha, nil, "Alpha Root", now)
	betaRoot := department("018bcfe5-6800-7000-8000-000000000112", beta, nil, "Beta Root", now)
	if err := store.CreateDepartment(ctx, alphaRoot); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateDepartment(ctx, betaRoot); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateDepartment(ctx, department("018bcfe5-6800-7000-8000-000000000113", alpha, nil, "Another Root", now)); err == nil {
		t.Fatal("second tenant root succeeded")
	}
	childID := organization.DepartmentID("018bcfe5-6800-7000-8000-000000000114")
	grandchildID := organization.DepartmentID("018bcfe5-6800-7000-8000-000000000115")
	if err := store.CreateDepartment(ctx, department(string(childID), alpha, &alphaRoot.ID, "Engineering", now)); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateDepartment(ctx, department(string(grandchildID), alpha, &childID, "Platform", now)); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateDepartment(ctx, department("018bcfe5-6800-7000-8000-000000000116", alpha, &alphaRoot.ID, " engineering ", now)); err == nil {
		t.Fatal("duplicate normalized sibling succeeded")
	}
	if err := store.CreateDepartment(ctx, department("018bcfe5-6800-7000-8000-000000000117", alpha, &betaRoot.ID, "Cross Tenant", now)); err == nil {
		t.Fatal("cross-tenant parent succeeded")
	}
	if err := store.MoveDepartment(ctx, alpha, childID, grandchildID, now); !errors.Is(err, organization.ErrCycle) {
		t.Fatalf("cycle move error = %v", err)
	}
	if err := store.MoveDepartment(ctx, alpha, childID, betaRoot.ID, now); !errors.Is(err, organization.ErrNotFound) {
		t.Fatalf("cross-tenant move error = %v", err)
	}
	if err := store.DeleteDepartment(ctx, alpha, alphaRoot.ID); !errors.Is(err, organization.ErrRootDepartment) {
		t.Fatalf("root delete error = %v", err)
	}
	if err := store.DeleteDepartment(ctx, alpha, childID); !errors.Is(err, organization.ErrInUse) {
		t.Fatalf("parent delete error = %v", err)
	}

	position := organization.Position{ID: "018bcfe5-6800-7000-8000-000000000121", TenantID: alpha, Name: "Engineer", NormalizedName: "engineer", CreatedAt: now}
	if err := store.CreatePosition(ctx, position); err != nil {
		t.Fatal(err)
	}
	betaPosition := organization.Position{ID: "018bcfe5-6800-7000-8000-000000000122", TenantID: beta, Name: "Engineer", NormalizedName: "engineer", CreatedAt: now}
	if err := store.CreatePosition(ctx, betaPosition); err != nil {
		t.Fatal(err)
	}
	if err := store.AssignUser(ctx, alpha, "018bcfe5-6800-7000-8000-000000000131", childID, &position.ID, now); err != nil {
		t.Fatal(err)
	}
	if err := store.AssignUser(ctx, alpha, "018bcfe5-6800-7000-8000-000000000131", childID, &betaPosition.ID, now); err == nil {
		t.Fatal("cross-tenant position assignment succeeded")
	}
	if err := store.AssignUser(ctx, beta, "018bcfe5-6800-7000-8000-000000000132", childID, nil, now); err == nil {
		t.Fatal("cross-tenant department assignment succeeded")
	}
	if err := store.DeleteDepartment(ctx, alpha, grandchildID); err != nil {
		t.Fatal(err)
	}
	departments, err := store.ListDepartments(ctx, alpha)
	if err != nil || len(departments) != 2 {
		t.Fatalf("alpha departments=%+v err=%v", departments, err)
	}
	positions, err := store.ListPositions(ctx, alpha)
	if err != nil || len(positions) != 1 || positions[0].ID != position.ID {
		t.Fatalf("alpha positions=%+v err=%v", positions, err)
	}
}

func department(id string, tenantID identity.TenantID, parentID *organization.DepartmentID, name string, now time.Time) organization.Department {
	return organization.Department{ID: organization.DepartmentID(id), TenantID: tenantID, ParentID: parentID, Name: strings.TrimSpace(name), NormalizedName: organization.NormalizeName(name), CreatedAt: now}
}

func seedIdentity(t *testing.T, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO modura.tenants (id, slug, display_name, status, created_at, updated_at) VALUES
('018bcfe5-6800-7000-8000-000000000101', 'org-alpha', 'Alpha', 'active', $1, $1),
('018bcfe5-6800-7000-8000-000000000102', 'org-beta', 'Beta', 'active', $1, $1)`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `
INSERT INTO modura.users (id, tenant_id, username, normalized_username, password_hash, status, created_at, updated_at) VALUES
('018bcfe5-6800-7000-8000-000000000131', '018bcfe5-6800-7000-8000-000000000101', 'alpha-user', 'alpha-user', 'hash', 'active', $1, $1),
('018bcfe5-6800-7000-8000-000000000132', '018bcfe5-6800-7000-8000-000000000102', 'beta-user', 'beta-user', 'hash', 'active', $1, $1)`, now); err != nil {
		t.Fatal(err)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("MODURA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MODURA_TEST_DATABASE_URL is not set")
	}
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(config.ConnConfig.Database, "_test") {
		t.Fatalf("refusing destructive integration setup for database %q: name must end in _test", config.ConnConfig.Database)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	lockConnection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lockConnection.Exec(context.Background(), "SELECT pg_advisory_lock(1297040469)"); err != nil {
		lockConnection.Release()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = lockConnection.Exec(context.Background(), "SELECT pg_advisory_unlock(1297040469)")
		lockConnection.Release()
	})
	if _, err := pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS modura CASCADE"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS modura CASCADE") })
	for _, name := range []string{"000001_initialize.up.sql", "000002_identity_foundation.up.sql", "000003_organization_foundation.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "..", "platform", "database", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(context.Background(), string(migration)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
	return pool
}
