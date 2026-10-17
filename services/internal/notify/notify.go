// Package notify consumes domain events from Kafka and delivers e-mails
// (OTP codes, e-tickets, cancellation receipts) over SMTP. SMS is simulated
// by logging.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

type Notifier struct {
	smtpAddr string
	from     string
	siteURL  string
}

func New() *Notifier {
	return &Notifier{
		smtpAddr: platform.Env("SMTP_ADDR", "localhost:1025"),
		from:     platform.Env("MAIL_FROM", "tickets@showbuzz.local"),
		siteURL:  platform.Env("SITE_URL", "http://localhost:8080"),
	}
}

// Run consumes until ctx is cancelled.
func (n *Notifier) Run(ctx context.Context) error {
	cl, err := platform.NewConsumer("notify", platform.TopicOTPRequested, platform.TopicBookingConfirmed, platform.TopicBookingCancelled)
	if err != nil {
		return err
	}
	defer cl.Close()
	for {
		fetches := cl.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}
		fetches.EachError(func(topic string, _ int32, err error) {
			slog.Warn("fetch error", "topic", topic, "err", err)
		})
		fetches.EachRecord(func(rec *kgo.Record) {
			if err := n.handle(rec); err != nil {
				slog.Error("handle event", "topic", rec.Topic, "err", err)
			}
		})
		if err := cl.CommitUncommittedOffsets(ctx); err != nil && ctx.Err() == nil {
			slog.Warn("commit offsets", "err", err)
		}
	}
}

func (n *Notifier) handle(rec *kgo.Record) error {
	switch rec.Topic {
	case platform.TopicOTPRequested:
		var e struct{ Identifier, Channel, OTP string }
		if err := json.Unmarshal(rec.Value, &e); err != nil {
			return err
		}
		if e.Channel == "sms" {
			slog.Info("sms sent (simulated)", "to", "+91"+e.Identifier, "text", e.OTP+" is your showbuzz OTP. Valid for 5 minutes.")
			return nil
		}
		return n.send(e.Identifier, e.OTP+" is your showbuzz verification code", otpTmpl, map[string]string{"OTP": e.OTP})
	case platform.TopicBookingConfirmed:
		var b booking
		if err := json.Unmarshal(rec.Value, &b); err != nil {
			return err
		}
		if b.ContactPhone != "" {
			slog.Info("sms sent (simulated)", "to", "+91"+b.ContactPhone, "text", fmt.Sprintf("Booking ID %s confirmed for %s.", b.ID, b.Snapshot.Title))
		}
		return n.send(b.ContactEmail, "Your tickets for "+b.Snapshot.Title+" | Booking ID "+b.ID, ticketTmpl, n.view(b, 0))
	case platform.TopicBookingCancelled:
		var e struct {
			Booking booking `json:"booking"`
			Refund  float64 `json:"refund"`
		}
		if err := json.Unmarshal(rec.Value, &e); err != nil {
			return err
		}
		return n.send(e.Booking.ContactEmail, "Booking "+e.Booking.ID+" cancelled", cancelTmpl, n.view(e.Booking, e.Refund))
	}
	return nil
}

type booking struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Qty      int    `json:"qty"`
	Snapshot struct {
		Title        string    `json:"title"`
		Certificate  string    `json:"certificate"`
		Language     string    `json:"language"`
		Format       string    `json:"format"`
		VenueName    string    `json:"venueName"`
		VenueAddress string    `json:"venueAddress"`
		ScreenName   string    `json:"screenName"`
		StartsAt     time.Time `json:"startsAt"`
		TierName     string    `json:"tierName"`
		Lines        []struct {
			AreaName string   `json:"areaName"`
			Seats    []string `json:"seats"`
		} `json:"lines"`
	} `json:"snapshot"`
	Fnb []struct {
		Name string `json:"name"`
		Qty  int    `json:"qty"`
	} `json:"fnb"`
	Pricing struct {
		Total float64 `json:"total"`
	} `json:"pricing"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
}

func (n *Notifier) view(b booking, refund float64) map[string]any {
	var seats []string
	for _, l := range b.Snapshot.Lines {
		if len(l.Seats) > 0 {
			seats = append(seats, l.AreaName+" - "+strings.Join(l.Seats, ", "))
		}
	}
	if b.Kind == "EVENT" {
		seats = []string{fmt.Sprintf("%d x %s", b.Qty, b.Snapshot.TierName)}
	}
	var fnb []string
	for _, f := range b.Fnb {
		fnb = append(fnb, fmt.Sprintf("%d x %s", f.Qty, f.Name))
	}
	return map[string]any{
		"B": b, "Seats": strings.Join(seats, " | "), "Fnb": strings.Join(fnb, ", "),
		"When":   b.Snapshot.StartsAt.In(platform.IST).Format("Mon, 02 Jan 2006 | 03:04 PM"),
		"Total":  fmt.Sprintf("%.2f", b.Pricing.Total),
		"Refund": fmt.Sprintf("%.2f", refund),
		"Link":   n.siteURL + "/booking/confirmation/" + b.ID,
	}
}

func (n *Notifier) send(to, subject string, tmpl *template.Template, data any) error {
	if to == "" {
		return nil
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}
	msg := "From: showbuzz <" + n.from + ">\r\nTo: " + to + "\r\nSubject: " + subject +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n" + body.String()
	if err := smtp.SendMail(n.smtpAddr, nil, n.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp: %w", err)
	}
	slog.Info("email sent", "to", to, "subject", subject)
	return nil
}

const shell = `<div style="font-family:Roboto,Arial,sans-serif;max-width:560px;margin:0 auto;background:#fff;border:1px solid #eee">
<div style="background:#333545;padding:16px 24px;color:#fff;font-size:22px;font-weight:700">show<span style="background:#F84464;padding:0 6px;margin-left:2px;border-radius:3px;transform:rotate(-6deg);display:inline-block">buzz</span></div>
<div style="padding:24px">%s</div>
<div style="background:#f5f5f5;padding:12px 24px;color:#999;font-size:11px">This is a demo application. No real tickets were sold.</div></div>`

var otpTmpl = template.Must(template.New("otp").Parse(fmt.Sprintf(shell, `
<p style="color:#333">Use the code below to sign in to showbuzz. It is valid for 5 minutes.</p>
<p style="font-size:34px;letter-spacing:10px;font-weight:700;color:#222;margin:24px 0">{{.OTP}}</p>
<p style="color:#999;font-size:12px">If you didn't request this, you can safely ignore this email.</p>`)))

var ticketTmpl = template.Must(template.New("ticket").Parse(fmt.Sprintf(shell, `
<p style="color:#1EA83C;font-weight:700;font-size:18px;margin:0 0 16px">Your booking is confirmed!</p>
<table style="width:100%%;border-collapse:collapse;color:#333;font-size:14px">
<tr><td colspan="2" style="font-size:20px;font-weight:700;padding-bottom:4px">{{.B.Snapshot.Title}}</td></tr>
<tr><td colspan="2" style="color:#666;padding-bottom:12px">{{.B.Snapshot.Certificate}} {{.B.Snapshot.Language}} {{.B.Snapshot.Format}}</td></tr>
<tr><td style="color:#999;padding:6px 0">When</td><td>{{.When}}</td></tr>
<tr><td style="color:#999;padding:6px 0">Where</td><td>{{.B.Snapshot.VenueName}}{{if .B.Snapshot.ScreenName}} · {{.B.Snapshot.ScreenName}}{{end}}</td></tr>
<tr><td style="color:#999;padding:6px 0">Tickets</td><td>{{.Seats}}</td></tr>
{{if .Fnb}}<tr><td style="color:#999;padding:6px 0">Food &amp; Beverages</td><td>{{.Fnb}}</td></tr>{{end}}
<tr><td style="color:#999;padding:6px 0">Booking ID</td><td style="font-weight:700;letter-spacing:2px">{{.B.ID}}</td></tr>
<tr><td style="color:#999;padding:6px 0">Amount paid</td><td style="font-weight:700">Rs. {{.Total}}</td></tr>
</table>
<p style="margin-top:24px"><a href="{{.Link}}" style="background:#F84464;color:#fff;padding:12px 20px;border-radius:8px;text-decoration:none;font-weight:700">View M-Ticket</a></p>`)))

var cancelTmpl = template.Must(template.New("cancel").Parse(fmt.Sprintf(shell, `
<p style="font-weight:700;font-size:18px;color:#333">Your booking {{.B.ID}} has been cancelled</p>
<p style="color:#333">{{.B.Snapshot.Title}} · {{.When}}<br>{{.B.Snapshot.VenueName}}</p>
<p style="color:#333">A refund of <b>Rs. {{.Refund}}</b> will be credited to your original payment method within 5-7 working days. Convenience fees are non-refundable.</p>`)))
