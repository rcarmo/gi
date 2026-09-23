-- Durable per-scope invalidations. Separate from the immutable v1 declarations.
-- A refresh captures requested_revision at Begin and acknowledges only that
-- value on successful commit. Newer requests remain pending across crashes.
CREATE TABLE workspace_index_invalidations (
    workspace_id INTEGER NOT NULL,
    scope TEXT NOT NULL,
    requested_revision INTEGER NOT NULL DEFAULT 0 CHECK(requested_revision >= 0),
    acknowledged_revision INTEGER NOT NULL DEFAULT 0
        CHECK(acknowledged_revision >= 0 AND acknowledged_revision <= requested_revision),
    PRIMARY KEY(workspace_id, scope),
    FOREIGN KEY(workspace_id, scope) REFERENCES workspace_index_scopes(workspace_id, scope) ON DELETE CASCADE
);
INSERT INTO workspace_index_invalidations(workspace_id,scope)
SELECT workspace_id,scope FROM workspace_index_scopes;
