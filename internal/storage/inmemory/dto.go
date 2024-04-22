package inmemory

import (
	"time"
)

type RoomAvailability struct {
	HotelID  string
	RoomType string
	Date     time.Time
	Quota    uint
}
