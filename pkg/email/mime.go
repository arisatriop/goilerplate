package email

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

// buildMIME renders msg as an RFC 5322 message: text only, or multipart/alternative with an
// HTML part. Bodies are quoted-printable so non-ASCII text and long lines survive every relay.
//
// The API drivers do not need it; they hand the parts to the provider, which builds the
// message itself.
func buildMIME(from, to *mail.Address, msg Message, date time.Time, messageID string) ([]byte, error) {
	var buf bytes.Buffer

	// mail.Address.String encodes a non-ASCII display name, and QEncoding encodes a non-ASCII
	// subject. validate has already refused line breaks, so neither can inject a header.
	writeHeader(&buf, "From", formatAddress(from))
	writeHeader(&buf, "To", formatAddress(to))
	writeHeader(&buf, "Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	writeHeader(&buf, "Date", date.Format(time.RFC1123Z))
	writeHeader(&buf, "Message-ID", messageID)
	writeHeader(&buf, "MIME-Version", "1.0")

	if msg.HTML == "" {
		writeHeader(&buf, "Content-Type", `text/plain; charset="utf-8"`)
		writeHeader(&buf, "Content-Transfer-Encoding", "quoted-printable")
		buf.WriteString("\r\n")
		if err := writeQuotedPrintable(&buf, msg.Text); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	parts := multipart.NewWriter(&buf)
	writeHeader(&buf, "Content-Type", mime.FormatMediaType("multipart/alternative", map[string]string{"boundary": parts.Boundary()}))
	buf.WriteString("\r\n")

	// Plain text first: in multipart/alternative the last part is the preferred one.
	for _, part := range []struct{ contentType, body string }{
		{`text/plain; charset="utf-8"`, msg.Text},
		{`text/html; charset="utf-8"`, msg.HTML},
	} {
		w, err := parts.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {part.contentType},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, fmt.Errorf("creating MIME part: %w", err)
		}
		if err := writeQuotedPrintable(w, part.body); err != nil {
			return nil, err
		}
	}

	if err := parts.Close(); err != nil {
		return nil, fmt.Errorf("closing MIME message: %w", err)
	}
	return buf.Bytes(), nil
}

func writeHeader(buf *bytes.Buffer, name, value string) {
	buf.WriteString(name)
	buf.WriteString(": ")
	buf.WriteString(value)
	buf.WriteString("\r\n")
}

func writeQuotedPrintable(w interface{ Write([]byte) (int, error) }, body string) error {
	qp := quotedprintable.NewWriter(w)
	if _, err := qp.Write([]byte(body)); err != nil {
		return fmt.Errorf("encoding body: %w", err)
	}
	if err := qp.Close(); err != nil {
		return fmt.Errorf("encoding body: %w", err)
	}
	return nil
}

// messageIDDomain is the part of the Message-ID after the @. The sender's own domain keeps the
// ID globally unique without the host name, which would leak into every email sent.
func messageIDDomain(from *mail.Address) string {
	if at := strings.LastIndexByte(from.Address, '@'); at >= 0 {
		return from.Address[at+1:]
	}
	return "localhost"
}
