-- owner: audit
ALTER TABLE modura.audit_events
    ADD COLUMN before_state jsonb,
    ADD COLUMN after_state jsonb;
