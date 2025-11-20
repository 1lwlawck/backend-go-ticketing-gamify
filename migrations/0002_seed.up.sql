INSERT INTO users (id, name, username, password_hash, role, badges, bio)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'Avery Admin', 'avery', '$2a$10$JMFl6zQzspGDZsBoBPXa8e7OJbRRqtD9h5pz50jd0vBXZTUVsgx2.', 'admin', ARRAY['Launch Pioneer'], 'Oversees delivery and quality.'),
  ('00000000-0000-0000-0000-000000000002', 'Devon Dev', 'devon', '$2a$10$JMFl6zQzspGDZsBoBPXa8e7OJbRRqtD9h5pz50jd0vBXZTUVsgx2.', 'developer', ARRAY['Bug Smasher'], 'Full-stack engineer.'),
  ('00000000-0000-0000-0000-000000000003', 'Parker PM', 'parker', '$2a$10$JMFl6zQzspGDZsBoBPXa8e7OJbRRqtD9h5pz50jd0vBXZTUVsgx2.', 'project_manager', ARRAY['Sprint Strategist'], 'Keeps everyone aligned.');

INSERT INTO projects (id, name, description, created_by)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Core Platform Refresh', 'Streamlining ticket visibility and improving gamification loops.', '00000000-0000-0000-0000-000000000001');

INSERT INTO project_members (project_id, user_id, member_role)
VALUES
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000001', 'lead'),
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000002', 'member'),
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000003', 'member');

INSERT INTO tickets (id, project_id, title, description, status, priority, type, reporter_id, assignee_id)
VALUES
  ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Kanban board shell', 'Implement drag & drop columns for project board.', 'todo', 'high', 'feature', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002');

INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days)
VALUES
  ('00000000-0000-0000-0000-000000000001', 120, 2, 200, 8, 3),
  ('00000000-0000-0000-0000-000000000002', 80, 1, 100, 4, 2),
  ('00000000-0000-0000-0000-000000000003', 60, 1, 100, 3, 1);
