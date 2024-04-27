package handlers

import (
	"fmt"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/gin-gonic/gin"
)

type httpError struct {
	code int
	text string
}

func (err httpError) Error() string {
	return err.text
}

const TimeLayout = time.RFC3339

func setHeaders(ctx *gin.Context) {
	ctx.SetAccepted(gin.MIMEJSON)
	ctx.Header("Content-Type", "application/json; charset=utf-8")
}

func toDay(timestamp time.Time) time.Time {
	dayRounded := timestamp.UTC().Round(time.Hour * 24).Day()
	return time.Date(timestamp.Year(), timestamp.Month(), dayRounded, 0, 0, 0, 0, time.UTC)
}

type TimeJSON struct {
	time.Time
}

func newTimeJSON(time time.Time) *TimeJSON {
	return &TimeJSON{
		Time: time,
	}
}

func (t *TimeJSON) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.Time.Format(TimeLayout))), nil
}

func (t *TimeJSON) UnmarshalJSON(bts []byte) error {
	var err error
	t.Time, err = time.Parse(TimeLayout, string(bts[1:len(bts)-1]))

	return err
}

func reservationStatusToResponse(status booking.ReservationStatus) string {
	if status == booking.CanceledReservationStatus {
		return "canceled"
	} else if status == booking.FinishedReservationStatus {
		return "finished"
	}
	return "in_progress"
}

func errorJSON(text string) gin.H {
	return gin.H{"error": text}
}
