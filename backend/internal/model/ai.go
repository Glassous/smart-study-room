package model

import (
	"encoding/json"
	"time"
)

type AIConversation struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIMessage struct {
	ID              int64              `json:"id"`
	ConversationID  int64              `json:"conversation_id"`
	Role            string             `json:"role"`
	Content         string             `json:"content"`
	ToolCalls       json.RawMessage    `json:"tool_calls,omitempty"`
	ToolCallID      string             `json:"tool_call_id,omitempty"`
	Status          string             `json:"status"`
	CreatedAt       time.Time          `json:"created_at"`
	Recommendations []AIRecommendation `json:"recommendations,omitempty"`
}

type AIChatRequest struct {
	ConversationID int64  `json:"conversation_id"`
	Message        string `json:"message" binding:"required,max=1000"`
}

type AIRecommendation struct {
	SeatID     int64  `json:"seat_id"`
	SeatNo     string `json:"seat_no"`
	RoomID     int64  `json:"room_id"`
	RoomName   string `json:"room_name"`
	Date       string `json:"date"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Zone       string `json:"zone"`
	HasPower   bool   `json:"has_power"`
	NearWindow bool   `json:"near_window"`
	Reason     string `json:"reason"`
	Booked     bool   `json:"booked"`
	Disabled   bool   `json:"disabled"`
}

type SeatOccupancyWindow struct {
	SeatID    int64  `json:"seat_id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
