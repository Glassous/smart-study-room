--
-- PostgreSQL database dump
--

\restrict zqn7yFt07px1uLrZ3Fq6u0sxbcfw0R7MkwSIAwu9RVgWrbpXI7E0eK2StSh1bdt

-- Dumped from database version 18.6
-- Dumped by pg_dump version 18.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: btree_gist; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;


--
-- Name: EXTENSION btree_gist; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION btree_gist IS 'support for indexing common datatypes in GiST';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: credit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.credit_logs (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    delta integer NOT NULL,
    reason character varying(200) NOT NULL,
    reservation_id bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE credit_logs; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.credit_logs IS '信用分变动流水(正为加分负为扣分)';


--
-- Name: credit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.credit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: credit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.credit_logs_id_seq OWNED BY public.credit_logs.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    type character varying(30) NOT NULL,
    title character varying(200) NOT NULL,
    content text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT notifications_type_check CHECK (((type)::text = ANY ((ARRAY['reservation_success'::character varying, 'checkin_reminder'::character varying, 'violation'::character varying, 'credit_change'::character varying, 'waitlist_promoted'::character varying, 'system'::character varying])::text[])))
);


--
-- Name: TABLE notifications; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.notifications IS '站内消息通知';


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: reservations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reservations (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    seat_id bigint NOT NULL,
    res_date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    source character varying(20) DEFAULT 'manual'::character varying NOT NULL,
    checkin_at timestamp with time zone,
    checkout_at timestamp with time zone,
    leave_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reservations_check CHECK ((end_time > start_time)),
    CONSTRAINT reservations_source_check CHECK (((source)::text = ANY ((ARRAY['manual'::character varying, 'auto'::character varying, 'waitlist'::character varying])::text[]))),
    CONSTRAINT reservations_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'checked_in'::character varying, 'temp_leave'::character varying, 'completed'::character varying, 'cancelled'::character varying, 'violation'::character varying])::text[])))
);


--
-- Name: TABLE reservations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.reservations IS '座位预约单,状态机: pending/checked_in/temp_leave/completed/cancelled/violation';


--
-- Name: COLUMN reservations.source; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.reservations.source IS '来源: manual手动选座 auto智能分配 waitlist候补递补';


--
-- Name: reservations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reservations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reservations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reservations_id_seq OWNED BY public.reservations.id;


--
-- Name: rooms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rooms (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    location character varying(200) DEFAULT ''::character varying NOT NULL,
    open_time time without time zone DEFAULT '08:00:00'::time without time zone NOT NULL,
    close_time time without time zone DEFAULT '22:00:00'::time without time zone NOT NULL,
    seat_rows integer DEFAULT 6 NOT NULL,
    seat_cols integer DEFAULT 8 NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rooms_check CHECK ((close_time > open_time)),
    CONSTRAINT rooms_seat_cols_check CHECK (((seat_cols >= 1) AND (seat_cols <= 30))),
    CONSTRAINT rooms_seat_rows_check CHECK (((seat_rows >= 1) AND (seat_rows <= 30)))
);


--
-- Name: TABLE rooms; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.rooms IS '自习室/学习空间';


--
-- Name: COLUMN rooms.seat_rows; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.rooms.seat_rows IS '座位平面图行数(批量生成座位用)';


--
-- Name: rooms_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.rooms_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: rooms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.rooms_id_seq OWNED BY public.rooms.id;


--
-- Name: seats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seats (
    id bigint NOT NULL,
    room_id bigint NOT NULL,
    seat_no character varying(20) NOT NULL,
    row_no integer NOT NULL,
    col_no integer NOT NULL,
    zone character varying(30) DEFAULT 'regular'::character varying NOT NULL,
    has_power boolean DEFAULT false NOT NULL,
    near_window boolean DEFAULT false NOT NULL,
    status character varying(20) DEFAULT 'available'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT seats_col_no_check CHECK ((col_no >= 1)),
    CONSTRAINT seats_row_no_check CHECK ((row_no >= 1)),
    CONSTRAINT seats_status_check CHECK (((status)::text = ANY ((ARRAY['available'::character varying, 'maintenance'::character varying, 'disabled'::character varying])::text[]))),
    CONSTRAINT seats_zone_check CHECK (((zone)::text = ANY ((ARRAY['quiet'::character varying, 'regular'::character varying, 'discussion'::character varying, 'computer'::character varying])::text[])))
);


--
-- Name: TABLE seats; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.seats IS '自习座位,含平面图坐标与属性';


--
-- Name: COLUMN seats.zone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.seats.zone IS '功能区域: quiet静音 regular普通 discussion研讨 computer机房';


--
-- Name: seats_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.seats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: seats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.seats_id_seq OWNED BY public.seats.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    username character varying(50) NOT NULL,
    password_hash character varying(100) NOT NULL,
    real_name character varying(50) DEFAULT ''::character varying NOT NULL,
    student_no character varying(20),
    role character varying(20) DEFAULT 'student'::character varying NOT NULL,
    credit_score integer DEFAULT 100 NOT NULL,
    credit_banned_until timestamp with time zone,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_credit_score_check CHECK (((credit_score >= 0) AND (credit_score <= 100))),
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['student'::character varying, 'admin'::character varying])::text[]))),
    CONSTRAINT users_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'disabled'::character varying])::text[])))
);


--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.users IS '系统用户(学生/管理员)';


--
-- Name: COLUMN users.credit_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.users.credit_score IS '信用分,初始100,违约扣分履约加分';


--
-- Name: COLUMN users.credit_banned_until; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.users.credit_banned_until IS '低信用禁约截止时间,为空表示不受限';


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: waitlist; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.waitlist (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    room_id bigint NOT NULL,
    res_date date NOT NULL,
    start_time time without time zone NOT NULL,
    end_time time without time zone NOT NULL,
    zone character varying(30),
    has_power boolean,
    near_window boolean,
    status character varying(20) DEFAULT 'waiting'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT waitlist_check CHECK ((end_time > start_time)),
    CONSTRAINT waitlist_status_check CHECK (((status)::text = ANY ((ARRAY['waiting'::character varying, 'promoted'::character varying, 'cancelled'::character varying, 'expired'::character varying])::text[])))
);


--
-- Name: TABLE waitlist; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.waitlist IS '满座候补队列,按 created_at 先后排队,空位释放自动递补';


--
-- Name: waitlist_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.waitlist_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: waitlist_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.waitlist_id_seq OWNED BY public.waitlist.id;


--
-- Name: credit_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credit_logs ALTER COLUMN id SET DEFAULT nextval('public.credit_logs_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: reservations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations ALTER COLUMN id SET DEFAULT nextval('public.reservations_id_seq'::regclass);


--
-- Name: rooms id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rooms ALTER COLUMN id SET DEFAULT nextval('public.rooms_id_seq'::regclass);


--
-- Name: seats id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seats ALTER COLUMN id SET DEFAULT nextval('public.seats_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: waitlist id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waitlist ALTER COLUMN id SET DEFAULT nextval('public.waitlist_id_seq'::regclass);


--
-- Name: credit_logs credit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credit_logs
    ADD CONSTRAINT credit_logs_pkey PRIMARY KEY (id);


--
-- Name: reservations excl_seat_overlap; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT excl_seat_overlap EXCLUDE USING gist (seat_id WITH =, res_date WITH =, tsrange((res_date + start_time), (res_date + end_time)) WITH &&) WHERE (((status)::text = ANY ((ARRAY['pending'::character varying, 'checked_in'::character varying, 'temp_leave'::character varying])::text[])));


--
-- Name: reservations excl_user_overlap; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT excl_user_overlap EXCLUDE USING gist (user_id WITH =, res_date WITH =, tsrange((res_date + start_time), (res_date + end_time)) WITH &&) WHERE (((status)::text = ANY ((ARRAY['pending'::character varying, 'checked_in'::character varying, 'temp_leave'::character varying])::text[])));


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: reservations reservations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (id);


--
-- Name: rooms rooms_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_name_key UNIQUE (name);


--
-- Name: rooms rooms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_pkey PRIMARY KEY (id);


--
-- Name: seats seats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seats
    ADD CONSTRAINT seats_pkey PRIMARY KEY (id);


--
-- Name: seats seats_room_id_row_no_col_no_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seats
    ADD CONSTRAINT seats_room_id_row_no_col_no_key UNIQUE (room_id, row_no, col_no);


--
-- Name: seats seats_room_id_seat_no_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seats
    ADD CONSTRAINT seats_room_id_seat_no_key UNIQUE (room_id, seat_no);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_student_no_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_student_no_key UNIQUE (student_no);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: waitlist waitlist_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waitlist
    ADD CONSTRAINT waitlist_pkey PRIMARY KEY (id);


--
-- Name: idx_credit_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_credit_user ON public.credit_logs USING btree (user_id, created_at DESC);


--
-- Name: idx_notify_user_unread; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notify_user_unread ON public.notifications USING btree (user_id, is_read, created_at DESC);


--
-- Name: idx_res_seat_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_res_seat_date ON public.reservations USING btree (seat_id, res_date);


--
-- Name: idx_res_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_res_status ON public.reservations USING btree (status);


--
-- Name: idx_res_user_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_res_user_date ON public.reservations USING btree (user_id, res_date);


--
-- Name: idx_seats_room; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seats_room ON public.seats USING btree (room_id);


--
-- Name: idx_seats_room_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seats_room_status ON public.seats USING btree (room_id, status);


--
-- Name: idx_wait_room_slot; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wait_room_slot ON public.waitlist USING btree (room_id, res_date, status, created_at);


--
-- Name: idx_wait_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wait_user ON public.waitlist USING btree (user_id);


--
-- Name: credit_logs credit_logs_reservation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credit_logs
    ADD CONSTRAINT credit_logs_reservation_id_fkey FOREIGN KEY (reservation_id) REFERENCES public.reservations(id) ON DELETE SET NULL;


--
-- Name: credit_logs credit_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credit_logs
    ADD CONSTRAINT credit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reservations reservations_seat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_seat_id_fkey FOREIGN KEY (seat_id) REFERENCES public.seats(id) ON DELETE CASCADE;


--
-- Name: reservations reservations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: seats seats_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seats
    ADD CONSTRAINT seats_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE;


--
-- Name: waitlist waitlist_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waitlist
    ADD CONSTRAINT waitlist_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE;


--
-- Name: waitlist waitlist_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.waitlist
    ADD CONSTRAINT waitlist_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict zqn7yFt07px1uLrZ3Fq6u0sxbcfw0R7MkwSIAwu9RVgWrbpXI7E0eK2StSh1bdt

