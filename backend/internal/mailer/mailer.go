package mailer

import (
	"elena-backend/internal/config"
	"fmt"
	"time"

	gomail "gopkg.in/gomail.v2"
)

type Mailer struct {
	cfg config.SMTPConfig
}

func New(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

type BookingData struct {
	FirstName       string
	AppointmentDate time.Time
	CancelURL       string
}

func (m *Mailer) SendBookingConfirmation(to string, d BookingData) error {
	date := d.AppointmentDate.Format("02.01.2006 в 15:04")

	subject := fmt.Sprintf("%s, вы записаны на приём %s к Елене", d.FirstName, date)

	text := fmt.Sprintf(
		"%s, вы записаны на приём %s к Елене.\n\n"+
			"Вы можете отменить запись по ссылке:\n%s",
		d.FirstName, date, d.CancelURL,
	)

	return m.send(to, subject, text)
}

type CancellationData struct {
	FirstName       string
	AppointmentDate time.Time
	Refund          bool
	DeadlineHours   int
}

func (m *Mailer) SendCancellationConfirmation(to string, d CancellationData) error {
	date := d.AppointmentDate.Format("02.01.2006 в 15:04")

	var refundNote string
	if d.Refund {
		refundNote = "Оплата будет возвращена в течение нескольких рабочих дней."
	} else {
		refundNote = fmt.Sprintf(
			"К сожалению, отмена произошла менее чем за %d ч. до приёма — средства не возвращаются.",
			d.DeadlineHours,
		)
	}

	subject := fmt.Sprintf("%s, вы отменили запись к Елене", d.FirstName)

	text := fmt.Sprintf(
		"Вы отменили запись на приём %s к Елене.\n\n%s",
		date, refundNote,
	)

	return m.send(to, subject, text)
}

func (m *Mailer) SendPaymentFailedNotice(to, firstName string) error {
	subject := "Не удалось провести оплату"
	text := fmt.Sprintf(
		"%s, к сожалению, оплата не прошла, и бронь времени была отменена.\n\n"+
			"Вы можете повторить попытку записи на сайте.",
		firstName,
	)
	return m.send(to, subject, text)
}

func (m *Mailer) send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	d := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.User, m.cfg.Password)
	d.SSL = m.cfg.Port == 465

	if err := d.DialAndSend(msg); err != nil {
		return fmt.Errorf("mailer: %w", err)
	}
	return nil
}

type StubMailer struct{}

func (s *StubMailer) SendBookingConfirmation(_ string, d BookingData) error {
	fmt.Printf("[STUB MAIL] booking: to=%s date=%s cancel=%s\n",
		"<email>", d.AppointmentDate.Format(time.RFC3339), d.CancelURL)
	return nil
}

func (s *StubMailer) SendCancellationConfirmation(_ string, d CancellationData) error {
	fmt.Printf("[STUB MAIL] cancellation: date=%s refund=%v deadline_hours=%d\n",
		d.AppointmentDate.Format(time.RFC3339), d.Refund, d.DeadlineHours)
	return nil
}

func (s *StubMailer) SendPaymentFailedNotice(to, firstName string) error {
	fmt.Printf("[STUB MAIL] payment failed: to=%s name=%s\n", to, firstName)
	return nil
}
