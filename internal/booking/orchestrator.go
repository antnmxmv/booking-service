package booking

import (
	"fmt"
	"sync"
	"time"
)

// DelayedQueue generalization for delayed message queue
type DelayedQueue interface {
	SendMessage(message string, delay time.Duration) error
	Subscribe() <-chan string
}

// ReservationOrchestrator does reservation lifecycle
// it needed to implement 2PC with list of services that implements service interface
type ReservationOrchestrator struct {
	repo    Repository
	jobs    []Job
	timeout time.Duration

	doneCh chan struct{}
}

// service represents some consistent operator
// it must not return error if request is duplicated, but must provide current request status

type Job interface {
	Name() ReservationStatus
	Run(*Reservation) (*bool, error)
	Cancel(*Reservation) (*bool, error)
	Subscribe() (<-chan JobResponse, error)
}

type JobResponse struct {
	ReservationID string
	IsSucceeded   bool
	UpdateData    func(r *Reservation)
	JobName       ReservationStatus
}

func NewReservationOrchestrator(
	r Repository,
	jobs ...Job,
) *ReservationOrchestrator {
	return &ReservationOrchestrator{
		repo:   r,
		jobs:   jobs,
		doneCh: make(chan struct{}),
	}
}

// execute runs all jobs from last reservation status
func (s *ReservationOrchestrator) execute(reservation *Reservation, skipCurrentJob bool) error {

	found := false
	if reservation.Status == CreatedReservationStatus {
		found = true
	}
	for _, job := range s.jobs {
		if !found && job.Name() == reservation.Status {
			found = true
			if skipCurrentJob {
				continue
			}
		} else if !found {
			continue
		}
		reservation.Status = ReservationStatus(job.Name())

		isSucceeded, err := job.Run(reservation)

		if err := s.repo.UpdateReservation(reservation); err != nil {
			return err
		}

		if err != nil {
			return err
		}

		if isSucceeded == nil {
			return nil
		}

		if !*isSucceeded {
			return fmt.Errorf("transaction failed on %s step", job.Name())
		}
	}

	reservation.Status = FinishedReservationStatus

	return nil
}

func (s *ReservationOrchestrator) rollback(reservation *Reservation, skipCurrentJob bool) error {
	found := false
	if reservation.Status == FinishedReservationStatus {
		found = true
	}
	for i := len(s.jobs) - 1; i >= 0; i-- {
		if !found && s.jobs[i].Name() == reservation.Status {
			found = true
			if skipCurrentJob {
				continue
			}
		} else if !found {
			continue
		}
		if _, err := s.jobs[i].Cancel(reservation); err != nil {
			reservation.Status = s.jobs[i].Name()

			return err
		}
	}
	reservation.Status = CreatedReservationStatus
	return nil
}

func (s *ReservationOrchestrator) consumePersistedData() error {
	notFinishedReservations, err := s.repo.GetNotFinishedReservations()
	if err != nil {
		return err
	}
	for _, r := range notFinishedReservations {
		s.execute(r, false)
	}

	return nil
}

// conumeIncomingChanges changes reservation status when payment order update happens
func (s *ReservationOrchestrator) consumeIncomingChanges() error {

	updatesCh := make(chan JobResponse)

	wg := sync.WaitGroup{}

	wg.Add(len(s.jobs))

	for _, job := range s.jobs {
		ch, err := job.Subscribe()
		if err != nil {
			return err
		}
		go func() {
			for update := range ch {
				updatesCh <- update
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(updatesCh)
	}()

	go func() {
		for {
			select {
			case update := <-updatesCh:
				r, err := s.repo.GetReservationByID(update.ReservationID)
				if err != nil {
					continue
				}
				if r.Status == update.JobName {
					update.UpdateData(r)
					if err := s.repo.UpdateReservation(r); err != nil {
						// nack or requeue
					}
					if update.IsSucceeded {
						_ = s.execute(r, true)
					}
				}

			case <-s.doneCh:
				return
			}
		}
	}()

	return nil
}

func (s *ReservationOrchestrator) run(timeout time.Duration) error {
	s.timeout = timeout

	if err := s.consumePersistedData(); err != nil {
		return err
	}

	if err := s.consumeIncomingChanges(); err != nil {
		return err
	}

	return nil
}

func (s *ReservationOrchestrator) stop() {
	close(s.doneCh)
}
