package handlers

import (
	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/payment"
)

func reservationModelToResponse(r *booking.Reservation) reservationResponse {
	return reservationResponse{
		ID:                    r.ID,
		HotelID:               r.HotelID,
		RoomTypes:             r.RoomTypes,
		PaymentType:           r.PaymentType,
		PaymentOrder:          r.PaymentOrder,
		PaymentRequestDetails: r.PaymentRequestDetails,
		StartDate:             newTimeJSON(r.StartDate),
		EndDate:               newTimeJSON(r.EndDate),
		Cost:                  r.Cost,
		Status:                reservationStatusToResponse(r.Status),
		AppliedDiscountIDs:    r.AppliedDiscountIDs,
		LastUpdateTime:        newTimeJSON(r.LastUpdateTime),
	}
}

type reservationResponse struct {
	ID                    string               `json:"id"`
	HotelID               string               `json:"hotel_id"`
	RoomTypes             booking.RoomsRequest `json:"rooms"`
	PaymentType           payment.SourceType   `json:"payment_type"`
	PaymentOrder          payment.Order        `json:"payment_order,omitempty"`
	PaymentRequestDetails payment.OrderDetails `json:"payment_request,omitempty"`
	StartDate             *TimeJSON            `json:"start_date"`
	EndDate               *TimeJSON            `json:"end_date"`
	Cost                  int                  `json:"cost"`
	Status                string               `json:"status"`
	AppliedDiscountIDs    []string             `json:"applied_discount_ids"`
	LastUpdateTime        *TimeJSON            `json:"last_update"`
}
