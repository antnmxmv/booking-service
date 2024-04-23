package handlers

import (
	"reflect"
	"testing"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/config"
	"github.com/antnmxmv/booking-service/internal/payment"
)

const (
	validHotelID = "123"
)

var paymentProviders = []payment.Source{
	payment.NewCashSource(),
	payment.NewCardSource(&config.Config{Payment: config.Payment{Card: config.Card{Timeout: time.Second * 5}}}),
}

func Test_reservationRequest_ToModel(t *testing.T) {
	now := &TimeJSON{time.Now()}

	defaultRoomsRequest := []roomsRequest{{RoomType: "lux", Count: 2}, {RoomType: "other", Count: 5}}
	defaulRoomsModel := booking.RoomsRequest{{RoomType: "lux", Count: 2}, {RoomType: "other", Count: 5}}
	tests := []struct {
		name    string
		req     createReservationRequest
		hotelID string
		want    *booking.ReservationRequest
		wantErr bool
	}{
		{
			name: "all good",
			req: createReservationRequest{
				RoomsRequest: defaultRoomsRequest,
				StartDate:    now,
				EndDate:      now,
				PaymentType:  "cash",
			},
			hotelID: validHotelID,
			want: &booking.ReservationRequest{
				HotelID:      validHotelID,
				RoomsRequest: defaulRoomsModel,
				StartDate:    toDay(now.Time),
				EndDate:      toDay(now.Time),
				PaymentType:  "cash",
			},
			wantErr: false,
		},
		{
			name: "wrong hotel_id format",
			req: createReservationRequest{
				RoomsRequest: defaultRoomsRequest,
				StartDate:    now,
				EndDate:      now,
				PaymentType:  "cash",
			},
			hotelID: "",
			want:    nil,
			wantErr: true,
		},

		{
			name: "merge room request",
			req: createReservationRequest{
				RoomsRequest: []roomsRequest{
					{RoomType: "1", Count: 1},
					{RoomType: "1", Count: 2},
					{RoomType: "2", Count: 1},
				},
				StartDate:   now,
				EndDate:     now,
				PaymentType: "cash",
			},
			hotelID: validHotelID,
			want: &booking.ReservationRequest{
				HotelID: validHotelID,
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "1", Count: 3},
					{RoomType: "2", Count: 1},
				},
				StartDate:   toDay(now.Time),
				EndDate:     toDay(now.Time),
				PaymentType: "cash",
			},
			wantErr: false,
		},
	}

	h := &reservationHandler{p: payment.NewPaymentProvider(paymentProviders...)}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &createReservationRequest{
				RoomsRequest:   tt.req.RoomsRequest,
				HotelID:        tt.hotelID,
				StartDate:      tt.req.StartDate,
				EndDate:        tt.req.EndDate,
				PaymentType:    "cash",
				PaymentDetails: []byte("{}"),
			}
			got, err := h.requestToModel(*r, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("reservationRequest.ToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(*got, *tt.want) {
					t.Errorf("reservationRequest.ToModel() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_reservationHandler_validate(t *testing.T) {
	tests := []struct {
		name string
		in   *booking.ReservationRequest
		out  error
	}{
		{
			name: "all good",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "lux", Count: 1},
				},
				PaymentType: "cash",
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			out: nil,
		},
		{
			name: "rooms count error (count = 0)",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "lux", Count: 0},
				},
				PaymentType: "cash",
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			out: roomsCountError,
		},
		{
			name: "rooms count error (len(arr) = 0)",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{},
				PaymentType:  "cash",
				StartDate:    time.Now(),
				EndDate:      time.Now(),
			},
			out: roomsCountError,
		},
		{
			name: "room type empty error",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "", Count: 1},
				},
				PaymentType: "cash",
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			out: roomTypeEmptyError,
		},
		{
			name: "wrong payment type",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "lux", Count: 1},
				},
				PaymentType: "wrong",
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			out: paymentTypeError,
		},
		{
			name: "dates order error",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "lux", Count: 1},
				},
				PaymentType: "cash",
				StartDate:   time.Now().Add(time.Hour),
				EndDate:     time.Now(),
			},
			out: datesOrderError,
		},
		{
			name: "old dates error",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "lux", Count: 1},
				},
				PaymentType: "cash",
				StartDate:   time.Time{},
				EndDate:     time.Now(),
			},
			out: wrongDatesError,
		},
		{
			name: "payment details parsing",
			in: &booking.ReservationRequest{
				RoomsRequest: []booking.RoomRequest{
					{RoomType: "lux", Count: 1},
				},
				PaymentType: "card",
				StartDate:   time.Now(),
				EndDate:     time.Now(),
			},
			out: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &reservationHandler{p: payment.NewPaymentProvider(paymentProviders...)}

			if err := h.validate(tt.in); err != tt.out {
				t.Errorf("reservationHandler.validate() error = %v, wantErr %v", err, tt.out)
			}
		})
	}
}
