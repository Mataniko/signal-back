package types

import (
	"encoding/xml"
	"log"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/xeals/signal-back/signal"
)

// SMSType is an SMS type as defined by the XML backup spec.
type SMSType uint64

// SMS types
const (
	SMSInvalid  SMSType = iota // 0
	SMSReceived                // 1
	SMSSent                    // 2
	SMSDraft                   // 3
	SMSOutbox                  // 4
	SMSFailed                  // 5
	SMSQueued                  // 6
)

// SMSes holds a set of MMS or SMS records.
type SMSes struct {
	XMLName xml.Name `xml:"smses"`
	Count   int      `xml:"count,attr"`
	MMS     []MMS    `xml:"mms"`
	SMS     []SMS    `xml:"sms"`
}

// SMS represents a Short Message Service record.
type SMS struct {
	XMLName       xml.Name `xml:"sms"`
	Protocol      *uint64  `xml:"protocol,attr"`       // optional
	Address       string   `xml:"address,attr"`        // required
	Date          string   `xml:"date,attr"`           // required
	Type          SMSType  `xml:"type,attr"`           // required
	Subject       *string  `xml:"subject,attr"`        // optional
	Body          string   `xml:"body,attr"`           // required
	TOA           *string  `xml:"toa,attr"`            // optional
	SCTOA         *string  `xml:"sc_toa,attr"`         // optional
	ServiceCenter *string  `xml:"service_center,attr"` // optional
	Read          uint64   `xml:"read,attr"`           // required
	Status        int64    `xml:"status,attr"`         // required
	Locked        *uint64  `xml:"locked,attr"`         // optional
	DateSent      *uint64  `xml:"date_sent,attr"`      // optional
	ReadableDate  *string  `xml:"readable_date,attr"`  // optional
	ContactName   *uint64  `xml:"contact_name,attr"`   // optional
}

// MMS represents a Multimedia Messaging Service record.
type MMS struct {
	XMLName      xml.Name  `xml:"mms"`
	Parts        []MMSPart `xml:"parts,attr"`
	TextOnly     *uint64   `xml:"text_only,attr"`     // optional
	Sub          *string   `xml:"sub,attr"`           // optional
	RetrSt       string    `xml:"retr_st,attr"`       // required
	Date         uint64    `xml:"date,attr"`          // required
	CtCls        string    `xml:"ct_cls,attr"`        // required
	SubCs        string    `xml:"sub_cs,attr"`        // required
	Read         uint64    `xml:"read,attr"`          // required
	CtL          string    `xml:"ct_l,attr"`          // required
	TrID         string    `xml:"tr_id,attr"`         // required
	St           string    `xml:"st,attr"`            // required
	MsgBox       uint64    `xml:"msg_box,attr"`       // required
	Address      uint64    `xml:"address,attr"`       // required
	MCls         string    `xml:"m_cls,attr"`         // required
	DTm          string    `xml:"d_tm,attr"`          // required
	ReadStatus   string    `xml:"read_status,attr"`   // required
	CtT          string    `xml:"ct_t,attr"`          // required
	RetrTxtCs    string    `xml:"retr_txt_cs,attr"`   // required
	DRpt         uint64    `xml:"d_rpt,attr"`         // required
	MId          string    `xml:"m_id,attr"`          // required
	DateSent     uint64    `xml:"date_sent,attr"`     // required
	Seen         uint64    `xml:"seen,attr"`          // required
	MType        uint64    `xml:"m_type,attr"`        // required
	V            uint64    `xml:"v,attr"`             // required
	Exp          string    `xml:"exp,attr"`           // required
	Pri          uint64    `xml:"pri,attr"`           // required
	Rr           uint64    `xml:"rr,attr"`            // required
	RespTxt      string    `xml:"resp_txt,attr"`      // required
	RptA         string    `xml:"rpt_a,attr"`         // required
	Locked       uint64    `xml:"locked,attr"`        // required
	RetrTxt      string    `xml:"retr_txt,attr"`      // required
	RespSt       string    `xml:"resp_st,attr"`       // required
	MSize        string    `xml:"m_size,attr"`        // required
	ReadableDate *string   `xml:"readable_date,attr"` // optional
	ContactName  *string   `xml:"contact_name,attr"`  // optional
}

// MMSPart holds a data blob for an MMS.
type MMSPart struct {
	XMLName xml.Name `xml:"part"`
	Seq     uint64   `xml:"seq,attr"`   // required
	Ct      uint64   `xml:"ct,attr"`    // required
	Name    string   `xml:"name,attr"`  // required
	ChSet   string   `xml:"chset,attr"` // required
	Cd      string   `xml:"cd,attr"`    // required
	Fn      string   `xml:"fn,attr"`    // required
	CID     string   `xml:"cid,attr"`   // required
	Cl      string   `xml:"cl,attr"`    // required
	CttS    string   `xml:"ctt_s,attr"` // required
	CttT    string   `xml:"ctt_t,attr"` // required
	Text    string   `xml:"text,attr"`  // required
	Data    *string  `xml:"data,attr"`  // optional
}

// NewSMSFromStatement constructs an XML SMS struct from a SQL statement.
func NewSMSFromStatement(stmt *signal.SqlStatement) (*SMS, error) {
	sms := StatementToSMS(stmt)
	if sms == nil {
		return nil, errors.Errorf("expected 22 columns for SMS, have %v", len(stmt.GetParameters()))
	}

	xml := SMS{
		Protocol:      &sms.Protocol,
		Subject:       sms.Subject,
		ServiceCenter: sms.ServiceCenter,
		Read:          sms.Read,
		Status:        int64(sms.Status),
		DateSent:      sms.DateSent,
		ReadableDate:  intToTime(sms.DateReceived),
	}

	if sms.Address != nil {
		xml.Address = *sms.Address
	}
	if sms.Type != nil {
		xml.Type = translateSMSType(*sms.Type)
	}
	if sms.Body != nil {
		xml.Body = *sms.Body
	}
	if sms.DateReceived != nil {
		xml.Date = strconv.FormatUint(*sms.DateReceived, 10)
	}
	if sms.Person != nil {
		xml.ContactName = sms.Person
	}

	return &xml, nil
}

func NewMMSFromStatement(stmt *signal.SqlStatement) (*MMS, error) {
	mms := StatementToMMS(stmt)
	if mms == nil {
		return nil, errors.Errorf("expected 42 columns for MMS, have %v", len(stmt.GetParameters()))
	}

	xml := MMS{}

	return &xml, nil
}

func intToTime(n *uint64) *string {
	if n == nil {
		return nil
	}
	unix := time.Unix(int64(*n)/1000, 0)
	t := unix.Format("Jan 02, 2006 3:04:05 PM")
	return &t
}

func translateSMSType(t uint64) SMSType {
	// Just get the lowest 5 bits, because everything else is masking.
	// https://github.com/signalapp/Signal-Android/blob/master/src/org/thoughtcrime/securesms/database/MmsSmsColumns.java
	v := uint8(t) & 0x1F

	switch v {
	// STANDARD
	case 1: // standard standard
		return SMSReceived
	case 2: // standard sent
		return SMSSent
	case 3: // standard draft
		return SMSDraft
	case 4: // standard outbox
		return SMSOutbox
	case 5: // standard failed
		return SMSFailed
	case 6: // standard queued
		return SMSQueued

		// SIGNAL
	case 20: // signal received
		return SMSReceived
	case 21: // signal outbox
		return SMSOutbox
	case 22: // signal sending
		return SMSQueued
	case 23: // signal sent
		return SMSSent
	case 24: // signal failed
		return SMSFailed
	case 25: // pending secure SMS fallback
		return SMSQueued
	case 26: // pending insecure SMS fallback
		return SMSQueued
	case 27: // signal draft
		return SMSDraft

	default:
		log.Fatalf("undefined SMS type: %#v\nplease report this issue, as well as (if possible) details about the SMS,\nsuch as whether it was sent, received, drafted, etc.\n", t)
		log.Fatalf("note that the output XML may not properly import to Signal\n")
		return SMSInvalid
	}
}
