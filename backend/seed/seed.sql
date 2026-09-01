-- ============================================================
-- 智能共享自习室预约系统 演示种子数据
-- 可重复执行: 先清空再插入
-- 默认口令: admin/admin123, stu01..stu40/123456
-- ============================================================
BEGIN;

TRUNCATE notifications, waitlist, credit_logs, reservations, seats, rooms, users
RESTART IDENTITY CASCADE;

-- ------------------------------------------------------------
-- 1. 用户: 1 管理员 + 40 学生
-- ------------------------------------------------------------
INSERT INTO users (username, password_hash, real_name, student_no, role) VALUES
('admin', '$2a$10$JfmOy/i2OKtmr.m5JOtcE.wAZSXMZ6BdxQowGZWbCgQdR6Z.EpMOW',
 '系统管理员', NULL, 'admin');

INSERT INTO users (username, password_hash, real_name, student_no, role)
SELECT 'stu' || lpad(g::text, 2, '0'),
       '$2a$10$n72zbvMj/t4UDXevDd4wm.g2SEX88XbqgBX5cSidx3Z5wG.441tN2',
       format('学生%s', lpad(g::text, 2, '0')),
       '2024' || lpad(g::text, 4, '0'),
       'student'
FROM generate_series(1, 40) g;

-- ------------------------------------------------------------
-- 2. 自习室 3 间
-- ------------------------------------------------------------
INSERT INTO rooms (name, location, open_time, close_time, seat_rows, seat_cols, description) VALUES
('1F-静音自习室', '图书馆一楼东侧', '08:00', '22:00', 6, 8,
 '全程静音自习区，适合深度学习与备考，禁止讨论与外放。'),
('2F-综合学习区', '图书馆二楼', '08:00', '22:00', 8, 10,
 '综合学习区，含静音排、普通排与研讨排，电源充足。'),
('3F-电脑研学区', '图书馆三楼西侧', '09:00', '21:00', 5, 6,
 '每座配备电源与千兆网口，适合编程与在线实验。');

-- ------------------------------------------------------------
-- 3. 座位批量生成(按房间行列, 含区域/电源/靠窗属性)
-- ------------------------------------------------------------
INSERT INTO seats (room_id, seat_no, row_no, col_no, zone, has_power, near_window)
SELECT r.id,
       chr(64 + rowg) || lpad(colg::text, 2, '0'),
       rowg,
       colg,
       CASE
         WHEN r.name LIKE '1F-%'  THEN 'quiet'
         WHEN r.name LIKE '3F-%'  THEN 'computer'
         WHEN rowg <= 2 THEN 'quiet'
         WHEN rowg <= 6 THEN 'regular'
         ELSE 'discussion'
       END,
       CASE WHEN r.name LIKE '3F-%' THEN TRUE ELSE (colg % 2 = 0) END,
       (colg = r.seat_cols)
FROM rooms r
CROSS JOIN generate_series(1, r.seat_rows) rowg
CROSS JOIN generate_series(1, r.seat_cols) colg;

-- ------------------------------------------------------------
-- 4. 近 14 天历史预约(供热力图/统计演示)
--    生成规则保证不触发排除约束:
--    a. 每位用户每天至多 1 条(用户时段不重叠)
--    b. 座位序号按 (uid*7 + d*23) mod 158 双射分配(座位不冲突)
-- ------------------------------------------------------------
INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time,
                          status, source, checkin_at, checkout_at, created_at)
SELECT t.uid, s.id, t.dt::date,
       make_time(t.st::int, 0, 0),
       make_time((t.st + t.dur)::int, 0, 0),
       CASE WHEN t.r100 < 88 THEN 'completed'
            WHEN t.r100 < 94 THEN 'violation'
            ELSE 'cancelled' END,
       'manual',
       CASE WHEN t.r100 < 88 THEN t.dt + make_interval(hours => t.st::int) + interval '4 minutes' END,
       CASE WHEN t.r100 < 88 THEN t.dt + make_interval(hours => (t.st + t.dur)::int) - interval '6 minutes' END,
       t.dt - interval '1 day' + make_interval(hours => (10 + (t.r100 % 8))::int)
FROM (
    SELECT u.id AS uid,
           (CURRENT_DATE - g.d)        AS dt,
           8 + ((u.id * 3 + g.d * 5) % 10)                 AS st,   -- 8..17 点开始
           LEAST(1 + ((u.id + g.d) % 3), 21 - 8)           AS dur,  -- 1..3 小时
           ((u.id * 31 + g.d * 17) % 100)                  AS r100,
           ((u.id * 7 + g.d * 23) % 158) + 1               AS seat_idx
    FROM users u
    CROSS JOIN generate_series(1, 14) AS g(d)
    WHERE u.role = 'student'
      AND ((u.id * 13 + g.d * 7) % 10) < 8                          -- 约 80% 出勤
) t
JOIN (
    SELECT id, row_number() OVER (ORDER BY id) AS rn FROM seats
) s ON s.rn = t.seat_idx;

-- ------------------------------------------------------------
-- 5. 今日/明日待签到预约(演示进行中状态)
-- ------------------------------------------------------------
INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time, status, source, created_at)
SELECT u.id, s.id, CURRENT_DATE, make_time(14, 0, 0), make_time(16, 0, 0), 'pending', 'manual', now() - interval '3 hours'
FROM users u, seats s
WHERE u.username = 'stu01' AND s.room_id = 1 AND s.seat_no = 'A01';

INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time, status, source, created_at)
SELECT u.id, s.id, CURRENT_DATE, make_time(19, 0, 0), make_time(21, 0, 0), 'pending', 'auto', now() - interval '2 hours'
FROM users u, seats s
WHERE u.username = 'stu02' AND s.room_id = 2 AND s.seat_no = 'B02';

INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time, status, source, created_at)
SELECT u.id, s.id, CURRENT_DATE + 1, make_time(9, 0, 0), make_time(12, 0, 0), 'pending', 'manual', now() - interval '1 hours'
FROM users u, seats s
WHERE u.username = 'stu03' AND s.room_id = 3 AND s.seat_no = 'A01';

INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time, status, source, created_at)
SELECT u.id, s.id, CURRENT_DATE + 1, make_time(14, 0, 0), make_time(18, 0, 0), 'pending', 'manual', now() - interval '1 hours'
FROM users u, seats s
WHERE u.username = 'stu04' AND s.room_id = 1 AND s.seat_no = 'C03';

-- 当前使用中(已签到)示例
INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time, status, source, checkin_at, created_at)
SELECT u.id, s.id, CURRENT_DATE, make_time(9, 0, 0), make_time(12, 0, 0), 'checked_in', 'manual', now() - interval '50 minutes', now() - interval '5 hours'
FROM users u, seats s
WHERE u.username = 'stu05' AND s.room_id = 2 AND s.seat_no = 'D05';

-- ------------------------------------------------------------
-- 6. 候补队列示例(明日 1F-静音自习室满座场景)
-- ------------------------------------------------------------
INSERT INTO waitlist (user_id, room_id, res_date, start_time, end_time, zone, has_power, near_window, status)
SELECT u.id, r.id, CURRENT_DATE + 1, '09:00', '12:00', 'quiet', TRUE, TRUE, 'waiting'
FROM users u, rooms r
WHERE u.username = 'stu06' AND r.name LIKE '1F-%';

INSERT INTO waitlist (user_id, room_id, res_date, start_time, end_time, zone, has_power, near_window, status)
SELECT u.id, r.id, CURRENT_DATE + 1, '09:00', '12:00', NULL, NULL, TRUE, 'waiting'
FROM users u, rooms r
WHERE u.username = 'stu07' AND r.name LIKE '1F-%';

-- ------------------------------------------------------------
-- 7. 信用流水与信用分(违约每单 -8 分)
-- ------------------------------------------------------------
INSERT INTO credit_logs (user_id, delta, reason, reservation_id, created_at)
SELECT r.user_id, -8, '超时未签到，系统判定违约', r.id,
       (r.res_date + interval '1 day') + make_interval(hours => 9)
FROM reservations r
WHERE r.status = 'violation';

UPDATE users u
SET credit_score = GREATEST(60, 100 - 8 *
    (SELECT count(*) FROM reservations r
     WHERE r.user_id = u.id AND r.status = 'violation'))
WHERE u.role = 'student';

COMMIT;
