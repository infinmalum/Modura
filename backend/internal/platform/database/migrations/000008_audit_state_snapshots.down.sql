-- owner: audit
ALTER TABLE modura.audit_events
    DROP COLUMN IF EXISTS after_state,
    DROP COLUMN IF EXISTS before_state;
