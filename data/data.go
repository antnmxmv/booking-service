package data

import (
	"time"

	"github.com/antnmxmv/booking-service/internal/storage/inmemory"
)

var hotelID = "aa500b05-98b6-4792-8378-9e46c1a1033d"

var RoomAvailability = []*inmemory.RoomAvailability{
	{HotelID: hotelID, RoomType: "lux", Date: date(2024, 6, 13), Quota: 0},
	{HotelID: hotelID, RoomType: "lux", Date: date(2024, 6, 14), Quota: 1},
	{HotelID: hotelID, RoomType: "eco", Date: date(2024, 6, 14), Quota: 2},
	{HotelID: hotelID, RoomType: "lux", Date: date(2024, 6, 15), Quota: 0},
	{HotelID: hotelID, RoomType: "lux", Date: date(2024, 6, 16), Quota: 0},
	{HotelID: hotelID, RoomType: "lux", Date: date(2024, 6, 17), Quota: 0},
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
