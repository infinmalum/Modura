package provisioning

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modura-dev/modura/backend/internal/modules/authorization"
	authorizationpostgres "github.com/modura-dev/modura/backend/internal/modules/authorization/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/identity"
	identitypostgres "github.com/modura-dev/modura/backend/internal/modules/identity/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/organization"
	organizationpostgres "github.com/modura-dev/modura/backend/internal/modules/organization/postgres"
	"github.com/modura-dev/modura/backend/internal/modules/platformadmin"
)

func TestProvisionIsAtomicAndIdempotent(t *testing.T) {
	pool := integrationPool(t)
	now := time.Unix(1_700_000_000, 0).UTC()
	key := []byte(strings.Repeat("k", 32))
	signer, err := identity.NewAccessTokenSigner("modura", "admin", "key", key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	identityService, err := identity.NewService(identitypostgres.New(pool), signer, identity.NewAccessTokenVerifier("modura", "admin", map[string][]byte{"key": key}, 0), identity.DefaultPasswordParameters(), time.Hour, func() time.Time { return now }, sequentialIDs(), func() (string, error) { return strings.Repeat("i", 43), nil })
	if err != nil {
		t.Fatal(err)
	}
	organizationService, err := organization.NewService(organizationpostgres.New(pool), func() time.Time { return now }, sequentialIDs())
	if err != nil {
		t.Fatal(err)
	}
	authorizationService, err := authorization.NewService(authorizationpostgres.New(pool))
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(pool, identityService, organizationService, authorizationService, func() time.Time { return now }, sequentialIDs(), func() (string, error) { return strings.Repeat("v", 43), nil }, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	request := Request{IdempotencyKey: "018bcfe5-6800-7000-8000-000000000301", Slug: " Acme ", DisplayName: "Acme", RootDepartmentName: "Acme Root", AdministratorUsername: "Admin", AdministratorEmail: "admin@example.com", Actor: platformadmin.Actor{AdministratorID: "018bcfe5-6800-7000-8000-000000000390", SessionID: "018bcfe5-6800-7000-8000-000000000391"}, Reason: "customer onboarding", CorrelationID: "request-provision-1"}
	first, err := service.Provision(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || first.TenantID == "" || first.AdministratorID == "" || first.InvitationToken == "" {
		t.Fatalf("first result = %+v", first)
	}
	second, err := service.Provision(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created || second.TenantID != first.TenantID || second.InvitationToken != "" {
		t.Fatalf("second result = %+v", second)
	}
	conflict := request
	conflict.DisplayName = "Different"
	if _, err := service.Provision(context.Background(), conflict); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	assertProvisionedGraph(t, pool, first)
	administrator := identity.Actor{TenantID: first.TenantID, UserID: first.AdministratorID, SessionID: "verified-session"}
	if err := authorizationService.Authorize(context.Background(), administrator, authorization.Permission{Resource: authorization.ResourceDepartments, Action: authorization.ActionDelete}); err != nil {
		t.Fatalf("tenant administrator permission denied: %v", err)
	}
	if err := authorizationService.Authorize(context.Background(), administrator, authorization.Permission{Resource: authorization.ResourcePositions, Action: authorization.ActionDelete}); !errors.Is(err, authorization.ErrDenied) {
		t.Fatalf("unregistered destructive permission error = %v", err)
	}
	nonAdministrator := identity.Actor{TenantID: first.TenantID, UserID: "018bcfe5-6800-7000-8000-000000000999", SessionID: "verified-session"}
	if err := authorizationService.Authorize(context.Background(), nonAdministrator, authorization.Permission{Resource: authorization.ResourceDepartments, Action: authorization.ActionRead}); !errors.Is(err, authorization.ErrDenied) {
		t.Fatalf("non-administrator permission error = %v", err)
	}

	duplicate := request
	duplicate.IdempotencyKey = "018bcfe5-6800-7000-8000-000000000302"
	if _, err := service.Provision(context.Background(), duplicate); err == nil {
		t.Fatal("duplicate tenant slug provisioning succeeded")
	}
	var tenants, audits int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM modura.tenants`).Scan(&tenants); err != nil {
		t.Fatal(err)
	}
	if tenants != 1 {
		t.Fatalf("tenant count after rollback = %d", tenants)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM modura.audit_events WHERE tenant_id = $1 AND action = 'tenant.provisioned'`, first.TenantID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 1 {
		t.Fatalf("provisioning audit count = %d", audits)
	}
}

func assertProvisionedGraph(t *testing.T, pool *pgxpool.Pool, result Result) {
	t.Helper()
	var status string
	var departments, roles, grants, assignments, invitations int
	query := `SELECT t.status,
    (SELECT count(*) FROM modura.departments d WHERE d.tenant_id = t.id),
    (SELECT count(*) FROM modura.roles r WHERE r.tenant_id = t.id AND r.code = 'tenant-admin' AND r.reserved),
    (SELECT count(*) FROM modura.user_roles ur WHERE ur.tenant_id = t.id),
    (SELECT count(*) FROM modura.user_organization uo WHERE uo.tenant_id = t.id),
    (SELECT count(*) FROM modura.auth_one_time_tokens tok WHERE tok.tenant_id = t.id AND tok.purpose = 'invitation')
FROM modura.tenants t WHERE t.id = $1`
	if err := pool.QueryRow(context.Background(), query, result.TenantID).Scan(&status, &departments, &roles, &grants, &assignments, &invitations); err != nil {
		t.Fatal(err)
	}
	if status != "active" || departments != 1 || roles != 1 || grants != 1 || assignments != 1 || invitations != 1 {
		t.Fatalf("status=%s departments=%d roles=%d grants=%d assignments=%d invitations=%d", status, departments, roles, grants, assignments, invitations)
	}
}

func sequentialIDs() func(time.Time) (string, error) {
	sequence := 0
	return func(time.Time) (string, error) {
		sequence++
		return fmt.Sprintf("018bcfe5-6800-7000-8000-%012d", 400+sequence), nil
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
	for _, name := range []string{"000001_initialize.up.sql", "000002_identity_foundation.up.sql", "000003_organization_foundation.up.sql", "000004_authorization_and_provisioning.up.sql", "000005_platform_identity.up.sql", "000006_platform_tenant_audit.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "platform", "database", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(context.Background(), string(migration)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
	return pool
}
