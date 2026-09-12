package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatsRepo 统计分析仓储(全部为只读聚合查询)
type StatsRepo struct {
	pool *pgxpool.Pool
}

func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{pool: pool}
}

// 占用类状态口径: completed / checked_in / temp_leave / violation
const occupiedSQL = `('completed', 'checked_in', 'temp_leave', 'violation')`

// SeatHourOccupancy 座位×小时占用: (seat_id, hour, occupied)
func (r *StatsRepo) SeatHourOccupancy(ctx context.Context, roomID int64, date string) ([]int64, []string, []bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id,
		       to_char(h.hour, 'HH24:00') AS hour,
		       EXISTS (
		           SELECT 1 FROM reservations rv
		           WHERE rv.seat_id = s.id
		             AND rv.res_date = $2::date
		             AND rv.status IN `+occupiedSQL+`
		             AND tsrange((rv.res_date + rv.start_time)::timestamp,
		                         (rv.res_date + rv.end_time)::timestamp)
		               && tsrange(h.hour, h.hour + interval '1 hour')
		       ) AS occupied
		FROM seats s
		CROSS JOIN LATERAL generate_series(
		     $2::date + (SELECT open_time FROM rooms WHERE id = $1),
		     $2::date + (SELECT close_time FROM rooms WHERE id = $1) - interval '1 hour',
		     interval '1 hour') AS h(hour)
		WHERE s.room_id = $1
		ORDER BY s.row_no, s.col_no, h.hour`, roomID, date)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	var seatIDs []int64
	var hours []string
	var flags []bool
	for rows.Next() {
		var id int64
		var hour string
		var occupied bool
		if err := rows.Scan(&id, &hour, &occupied); err != nil {
			return nil, nil, nil, err
		}
		seatIDs = append(seatIDs, id)
		hours = append(hours, hour)
		flags = append(flags, occupied)
	}
	return seatIDs, hours, flags, rows.Err()
}

// UtilizationByDate 近 N 天整体利用率(占用座位小时 / 可用座位小时)
func (r *StatsRepo) UtilizationByDate(ctx context.Context, days int) ([]string, []float64, []int, error) {
	rows, err := r.pool.Query(ctx, `
		WITH span AS (
		    SELECT generate_series(CURRENT_DATE - $1::int, CURRENT_DATE - 1, interval '1 day')::date AS d
		), cap AS (
		    SELECT span.d,
		           sum(( EXTRACT(EPOCH FROM (r.close_time - r.open_time)) / 3600.0 )
		               * r2.seat_cnt) AS capacity
		    FROM span
		    CROSS JOIN rooms r
		    CROSS JOIN LATERAL (SELECT count(*) AS seat_cnt FROM seats WHERE room_id = r.id) r2
		    GROUP BY span.d
		), used AS (
		    SELECT rv.res_date AS d,
		           sum(EXTRACT(EPOCH FROM (least(rv.end_time, r.close_time) - greatest(rv.start_time, r.open_time))) / 3600.0) AS hours,
		           count(*) AS cnt
		    FROM reservations rv
		    JOIN seats s ON s.id = rv.seat_id
		    JOIN rooms r ON r.id = s.room_id
		    WHERE rv.res_date >= CURRENT_DATE - $1::int
		      AND rv.res_date < CURRENT_DATE
		      AND rv.status IN `+occupiedSQL+`
		    GROUP BY rv.res_date
		)
		SELECT to_char(cap.d, 'YYYY-MM-DD'),
		       COALESCE(used.hours, 0) / NULLIF(cap.capacity, 0),
		       COALESCE(used.cnt, 0)
		FROM cap LEFT JOIN used ON used.d = cap.d
		ORDER BY cap.d`, days)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	var dates []string
	var utils []float64
	var counts []int
	for rows.Next() {
		var d string
		var u float64
		var c int
		if err := rows.Scan(&d, &u, &c); err != nil {
			return nil, nil, nil, err
		}
		dates = append(dates, d)
		utils = append(utils, u)
		counts = append(counts, c)
	}
	return dates, utils, counts, rows.Err()
}

// PeakHours 近 N 天各整点开始的预约次数
func (r *StatsRepo) PeakHours(ctx context.Context, days int) ([]string, []int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT to_char(h.hour, 'HH24:00'),
		       (SELECT count(*) FROM reservations rv
		        WHERE rv.res_date >= CURRENT_DATE - $1::int
		          AND rv.res_date < CURRENT_DATE
		          AND rv.status IN `+occupiedSQL+`
		          AND extract(hour from rv.start_time) = extract(hour from h.hour))
		FROM generate_series(timestamp '2000-01-01 07:00', timestamp '2000-01-01 22:00', interval '1 hour') AS h(hour)`,
		days)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var hours []string
	var counts []int
	for rows.Next() {
		var h string
		var c int
		if err := rows.Scan(&h, &c); err != nil {
			return nil, nil, err
		}
		hours = append(hours, h)
		counts = append(counts, c)
	}
	return hours, counts, rows.Err()
}

// TopSeats 近 N 天热门座位(按占用小时数)
func (r *StatsRepo) TopSeats(ctx context.Context, days int, limit int) ([]string, []string, []float64, []int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.seat_no, r.name,
		       round(sum(EXTRACT(EPOCH FROM (rv.end_time - rv.start_time)) / 3600.0)::numeric, 1)::float8,
		       count(*)
		FROM reservations rv
		JOIN seats s ON s.id = rv.seat_id
		JOIN rooms r ON r.id = s.room_id
		WHERE rv.res_date >= CURRENT_DATE - $1::int
		  AND rv.res_date < CURRENT_DATE
		  AND rv.status IN `+occupiedSQL+`
		GROUP BY s.seat_no, r.name
		ORDER BY 3 DESC
		LIMIT $2`, days, limit)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer rows.Close()
	var seatNos, roomNames []string
	var hours []float64
	var counts []int
	for rows.Next() {
		var no, rn string
		var h float64
		var c int
		if err := rows.Scan(&no, &rn, &h, &c); err != nil {
			return nil, nil, nil, nil, err
		}
		seatNos = append(seatNos, no)
		roomNames = append(roomNames, rn)
		hours = append(hours, h)
		counts = append(counts, c)
	}
	return seatNos, roomNames, hours, counts, rows.Err()
}

// TodaySummary 今日概览: 总座位/今日活跃/当前使用中/今日利用率分子分母
func (r *StatsRepo) TodaySummary(ctx context.Context) (totalSeats, activeToday, inUseNow int, usedHours, capacityHours float64, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT
		    (SELECT count(*) FROM seats WHERE status = 'available'),
		    (SELECT count(*) FROM reservations
		     WHERE res_date = CURRENT_DATE AND status <> 'cancelled'),
		    (SELECT count(*) FROM reservations
		     WHERE res_date = CURRENT_DATE AND status IN ('checked_in', 'temp_leave')),
		    COALESCE((SELECT sum(EXTRACT(EPOCH FROM (rv.end_time - rv.start_time)) / 3600.0)
		     FROM reservations rv
		     WHERE rv.res_date = CURRENT_DATE AND rv.status IN `+occupiedSQL+`), 0),
		    COALESCE((SELECT sum(EXTRACT(EPOCH FROM (r.close_time - r.open_time)) / 3600.0 * c.seat_cnt)
		     FROM rooms r
		     JOIN LATERAL (SELECT count(*) AS seat_cnt FROM seats s
		                   WHERE s.room_id = r.id AND s.status = 'available') c ON TRUE), 0)`).
		Scan(&totalSeats, &activeToday, &inUseNow, &usedHours, &capacityHours)
	return
}
