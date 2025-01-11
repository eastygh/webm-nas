INSERT INTO users (name, email, password, created_at) VALUES
    ('admin', 'admin@weave.com', '$2a$10$5whQjJqSdL18PrEP.z/gZOubMKhFB38K0CvHWdnaQodb/H3yeG4J2', time()),
                                                          ('demo', 'admin@weave.com', '$2a$10$5whQjJqSdL18PrEP.z/gZOubMKhFB38K0CvHWdnaQodb/H3yeG4J2', time());

INSERT INTO groups (name, kind, describe, created_at) VALUES
    ('root', 'system', 'weave system group', time()),
                                                          ('system:authenticated', 'system', 'system group contains all authenticated user', time()),
                                                          ('system:unauthenticated', 'system', 'system group contains all unauthenticated user', time())  ON CONFLICT DO NOTHING;
INSERT INTO user_groups (group_id, user_id)
SELECT  g.id, u.id FROM users AS u, groups AS g
WHERE (u.name='admin' AND g.name='root') ON CONFLICT DO NOTHING;

INSERT INTO user_groups (group_id, user_id)
SELECT  g.id, u.id FROM users AS u, groups AS g
WHERE u.name='demo' ON CONFLICT DO NOTHING;

INSERT INTO roles (name, scope, rules) VALUES
                                           ('cluster-admin', 'cluster', '[{"resource": "*", "operation": "*"}]'),
                                           ('authenticated', 'cluster', '[{"resource": "users", "operation": "*"},{"resource": "auth", "operation": "*"}]'),
                                           ('unauthenticated', 'cluster', '[{"resource": "auth", "operation": "create"}]');


INSERT INTO group_roles (group_id, role_id) VALUES
                                                ((SELECT id FROM groups WHERE name = 'root'), (SELECT id FROM roles WHERE name = 'cluster-admin')),
                                                ((SELECT id FROM groups WHERE name = 'system:authenticated'), (SELECT id FROM roles WHERE name = 'authenticated')),
                                                ((SELECT id FROM groups WHERE name = 'system:unauthenticated'), (SELECT id FROM roles WHERE name = 'unauthenticated'));

