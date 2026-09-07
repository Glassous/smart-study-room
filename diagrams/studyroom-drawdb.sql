-- drawDB PostgreSQL import: verified against live schema on 2026-09-07.
-- Excludes btree_gist extension and EXCLUDE constraints; see studyroom-schema-full.sql.
-- ============================================================
-- 智能共享自习室预约系统 数据库初始化脚本
-- 数据库: PostgreSQL 18 / studyroom
-- 说明: 依次创建 7 张核心表、约束与索引
-- ============================================================

-- btree_gist 扩展: 支持排除约束(时段冲突检测)
-- ------------------------------------------------------------
-- 1. users 用户表
-- ------------------------------------------------------------
CREATE TABLE users (
    id                BIGSERIAL PRIMARY KEY,
    username          VARCHAR(50)  NOT NULL UNIQUE,
    password_hash     VARCHAR(100) NOT NULL,
    real_name         VARCHAR(50)  NOT NULL DEFAULT '',
    student_no        VARCHAR(20)  UNIQUE,
    role              VARCHAR(20)  NOT NULL DEFAULT 'student'
                      CHECK (role IN ('student', 'admin')),
    credit_score      INTEGER      NOT NULL DEFAULT 100
                      CHECK (credit_score BETWEEN 0 AND 100),
    credit_banned_until TIMESTAMPTZ,
    status            VARCHAR(20)  NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'disabled')),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE users IS '系统用户(学生/管理员)';
COMMENT ON COLUMN users.credit_score IS '信用分,初始100,违约扣分履约加分';
COMMENT ON COLUMN users.credit_banned_until IS '低信用禁约截止时间,为空表示不受限';

-- ------------------------------------------------------------
-- 2. rooms 自习室表
-- ------------------------------------------------------------
CREATE TABLE rooms (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    location    VARCHAR(200) NOT NULL DEFAULT '',
    open_time   TIME         NOT NULL DEFAULT '08:00',
    close_time  TIME         NOT NULL DEFAULT '22:00',
    seat_rows   INTEGER      NOT NULL DEFAULT 6
                CHECK (seat_rows BETWEEN 1 AND 30),
    seat_cols   INTEGER      NOT NULL DEFAULT 8
                CHECK (seat_cols BETWEEN 1 AND 30),
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (close_time > open_time)
);

COMMENT ON TABLE rooms IS '自习室/学习空间';
COMMENT ON COLUMN rooms.seat_rows IS '座位平面图行数(批量生成座位用)';

-- ------------------------------------------------------------
-- 3. seats 座位表
-- ------------------------------------------------------------
CREATE TABLE seats (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    seat_no     VARCHAR(20) NOT NULL,
    row_no      INTEGER NOT NULL CHECK (row_no >= 1),
    col_no      INTEGER NOT NULL CHECK (col_no >= 1),
    zone        VARCHAR(30) NOT NULL DEFAULT 'regular'
                CHECK (zone IN ('quiet', 'regular', 'discussion', 'computer')),
    has_power   BOOLEAN NOT NULL DEFAULT FALSE,
    near_window BOOLEAN NOT NULL DEFAULT FALSE,
    status      VARCHAR(20) NOT NULL DEFAULT 'available'
                CHECK (status IN ('available', 'maintenance', 'disabled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (room_id, seat_no),
    UNIQUE (room_id, row_no, col_no)
);

CREATE INDEX idx_seats_room ON seats(room_id);
CREATE INDEX idx_seats_room_status ON seats(room_id, status);

COMMENT ON TABLE seats IS '自习座位,含平面图坐标与属性';
COMMENT ON COLUMN seats.zone IS '功能区域: quiet静音 regular普通 discussion研讨 computer机房';

-- ------------------------------------------------------------
-- 4. reservations 预约表(核心)
-- 状态机: pending -> checked_in -> completed
--                -> temp_leave -> checked_in
--   pending 超时未签到 -> violation
--   pending/checked_in 手动取消 -> cancelled
-- ------------------------------------------------------------
CREATE TABLE reservations (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seat_id     BIGINT NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    res_date    DATE   NOT NULL,
    start_time  TIME   NOT NULL,
    end_time    TIME   NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'checked_in', 'temp_leave',
                                  'completed', 'cancelled', 'violation')),
    source      VARCHAR(20) NOT NULL DEFAULT 'manual'
                CHECK (source IN ('manual', 'auto', 'waitlist')),
    checkin_at  TIMESTAMPTZ,
    checkout_at TIMESTAMPTZ,
    leave_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_res_seat_date ON reservations(seat_id, res_date);
CREATE INDEX idx_res_user_date ON reservations(user_id, res_date);
CREATE INDEX idx_res_status ON reservations(status);

-- 同一座位同一天,活跃预约时段不得重叠(数据库层兜底)
-- 同一用户同一天,活跃预约时段不得重叠(防止一人多占)
COMMENT ON TABLE reservations IS '座位预约单,状态机: pending/checked_in/temp_leave/completed/cancelled/violation';
COMMENT ON COLUMN reservations.source IS '来源: manual手动选座 auto智能分配 waitlist候补递补';

-- ------------------------------------------------------------
-- 5. credit_logs 信用流水表
-- ------------------------------------------------------------
CREATE TABLE credit_logs (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delta          INTEGER NOT NULL,
    reason         VARCHAR(200) NOT NULL,
    reservation_id BIGINT REFERENCES reservations(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_credit_user ON credit_logs(user_id, created_at DESC);

COMMENT ON TABLE credit_logs IS '信用分变动流水(正为加分负为扣分)';

-- ------------------------------------------------------------
-- 6. waitlist 候补队列表
-- ------------------------------------------------------------
CREATE TABLE waitlist (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id     BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    res_date    DATE   NOT NULL,
    start_time  TIME   NOT NULL,
    end_time    TIME   NOT NULL,
    zone        VARCHAR(30),
    has_power   BOOLEAN,
    near_window BOOLEAN,
    status      VARCHAR(20) NOT NULL DEFAULT 'waiting'
                CHECK (status IN ('waiting', 'promoted', 'cancelled', 'expired')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_wait_room_slot ON waitlist(room_id, res_date, status, created_at);
CREATE INDEX idx_wait_user ON waitlist(user_id);

COMMENT ON TABLE waitlist IS '满座候补队列,按 created_at 先后排队,空位释放自动递补';

-- ------------------------------------------------------------
-- 7. notifications 消息通知表
-- ------------------------------------------------------------
CREATE TABLE notifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(30) NOT NULL
               CHECK (type IN ('reservation_success', 'checkin_reminder', 'violation',
                               'credit_change', 'waitlist_promoted', 'system')),
    title      VARCHAR(200) NOT NULL,
    content    TEXT NOT NULL,
    is_read    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notify_user_unread ON notifications(user_id, is_read, created_at DESC);

COMMENT ON TABLE notifications IS '站内消息通知';
