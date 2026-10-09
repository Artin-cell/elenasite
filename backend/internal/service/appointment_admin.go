package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"elena-backend/internal/mailer"
	"elena-backend/internal/models"
	"elena-backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrAlreadyCancelled = errors.New("запись уже отменена")
	ErrCannotCancel     = errors.New("эту запись нельзя отменить: приём уже завершён")
	ErrNothingToRefund  = errors.New("по этой записи нет оплаты, возвращать нечего")
	ErrManualPayment    = errors.New("оплата внесена вручную, вернуть её через ЮKassa нельзя; отмените запись без возврата")
	ErrRefundFailed     = errors.New("не удалось оформить возврат в ЮKassa, запись не отменена")
	ErrStatusChanged    = errors.New("статус записи только что изменился, обновите список")
)

type AdminCancelResult struct {
	Refunded     bool
	RefundStatus string
}

// AdminCancel отменяет запись по решению специалиста.
// refund=true: сначала возврат денег через API ЮKassa, и только если он принят, запись отменяется.
// refund=false: запись отменяется, деньги не трогаем.
func (s *AppointmentService) AdminCancel(ctx context.Context, apptID uuid.UUID, refund bool) (*AdminCancelResult, error) {
	appt, err := s.apptRepo.GetByIDWithRelations(ctx, apptID)
	if err != nil {
		return nil, err
	}

	switch appt.Status {
	case models.StatusCancelled:
		return nil, ErrAlreadyCancelled
	case models.StatusPending, models.StatusConfirmed:
	default:
		return nil, ErrCannotCancel
	}

	result := &AdminCancelResult{}

	if refund {
		if appt.PaymentStatus != models.PaymentStatusPaid {
			return nil, ErrNothingToRefund
		}
		if appt.PaymentID == "" {
			return nil, ErrManualPayment
		}
		info, err := s.payment.Refund(ctx, appt.PaymentID, appt.AmountKopeks)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRefundFailed, err)
		}
		if info.Status != "succeeded" && info.Status != "pending" {
			return nil, fmt.Errorf("%w: статус возврата %q", ErrRefundFailed, info.Status)
		}
		result.Refunded = true
		result.RefundStatus = info.Status
	}

	if err := s.cancelAppointmentStatus(ctx, appt); err != nil {
		if result.Refunded {
			log.Printf("admin cancel: возврат по записи %s принят (%s), но отменить запись не удалось: %v",
				apptID, result.RefundStatus, err)
		}
		return nil, err
	}

	// Возврат прошёл сразу: фиксируем статус и пишем клиенту. Если возврат ещё «в обработке»,
	// статус и письмо обработает вебхук refund.succeeded.
	if result.Refunded && result.RefundStatus == "succeeded" {
		flipped, err := s.apptRepo.MarkRefunded(ctx, apptID)
		if err != nil {
			log.Printf("admin cancel: не удалось пометить возврат по записи %s: %v", apptID, err)
		} else if flipped {
			if err := s.mailer.SendCancellationConfirmation(appt.Client.Email, mailer.CancellationData{
				FirstName:       appt.Client.FirstName,
				AppointmentDate: appt.StartsAt,
				Refund:          true,
				DeadlineHours:   s.cancellationDeadlineHours,
				StaffInitiated:  true,
			}); err != nil {
				log.Printf("admin cancel: письмо клиенту по записи %s не отправлено: %v", apptID, err)
			}
		}
	}

	return result, nil
}

func (s *AppointmentService) cancelAppointmentStatus(ctx context.Context, appt *models.Appointment) error {
	// Неоплаченная онлайн-запись: освобождаем слот и помечаем платёж как неуспешный.
	if appt.Status == models.StatusPending && appt.PaymentStatus == models.PaymentStatusPending {
		changed, err := s.apptRepo.CancelPending(ctx, appt.ID)
		if err != nil {
			return err
		}
		if !changed {
			return ErrStatusChanged
		}
		return nil
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		err := s.apptRepo.UpdateStatus(ctx, appt.ID, models.StatusCancelled)
		if err == nil {
			return nil
		}
		if errors.Is(err, repository.ErrInvalidStatusTransition) {
			return ErrStatusChanged
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
	}
	return lastErr
}
