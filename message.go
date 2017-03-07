package email

import (
	"bytes"
	"crypto/sha256"
	"errors"
	htpl "html/template"
	"io/ioutil"
	"mime"
	"path/filepath"
	"strconv"
	"sync"
	ttpl "text/template"
	"time"
)

// CTE represents a "Content-Transfer-Encoding" method identifier.
type CTE byte

const (
	// AutoCTE leaves it up to the package to determine CTE
	AutoCTE CTE = iota
	// QuotedPrintable indicates "quoted-printable" CTE
	QuotedPrintable
	// Base64 indicates "base64" CTE
	Base64
)

var (
	now = time.Now
)

// Message represents all the information necessary for composing an email message with optional
// external data, and sending it via a Sender.
type Message struct {
	sync.RWMutex
	domain        []byte
	subject       []byte
	subjectTplSrc string
	subjectTpl    *ttpl.Template
	sender        *Sender
	from, replyTo *Address
	to, cc, bcc   addrList
	parts         []*part
	text, html    *part
	attachments   []*attachment
	errors        []error
	prepared      bool
}

// Domain sets the domain portion of the generated message Id.
//
// If not specified, the domain is extracted from the sender email address - which is
// the right choice for most applications.
func (m *Message) Domain(domain string) *Message {
	m.Lock()
	defer m.Unlock()
	m.domain = []byte(domain)
	return m
}

func (m *Message) setSender(s *Sender) *Message {
	m.Lock()
	defer m.Unlock()
	m.sender = s
	return m
}

// Subject sets the text for the subject of the message.
func (m *Message) Subject(subject string) *Message {
	m.Lock()
	defer m.Unlock()
	m.subject = []byte(subject)
	return m
}

// SubjectTemplate sets a template for the subject of the message.
func (m *Message) SubjectTemplate(tpl string) *Message {
	var (
		t   *ttpl.Template
		err error
	)
	if tpl != "" {
		t, err = ttpl.New("").Parse(tpl)
		if err != nil {
			m.errors = append(m.errors, errors.New("invalid subject template:\n"+tpl+"\nerror: "+err.Error()))
			return m
		}
	}
	m.Lock()
	defer m.Unlock()
	m.subjectTplSrc = tpl
	m.subjectTpl = t
	return m
}

// From sets the From: email address.
func (m *Message) From(addr *Address) *Message {
	if addr != nil && !SeemsValidAddr(addr.Addr) {
		addr = nil
	}
	m.Lock()
	defer m.Unlock()
	m.from = addr
	return m
}

// To sets the To: email address(es). Last call overrides any previous calls, replacing rather than
// adding to the list.
func (m *Message) To(addr ...*Address) *Message {
	lst := make(addrList, 0, len(addr))
	for _, a := range addr {
		if a != nil && SeemsValidAddr(a.Addr) {
			lst = append(lst, a)
		}
	}
	m.Lock()
	defer m.Unlock()
	m.to = lst
	return m
}

// Cc sets the (optional) Cc: email addresses. Last call overrides any previous calls, replacing rather than
// adding to the list.
func (m *Message) Cc(addr ...*Address) *Message {
	lst := make(addrList, 0, len(addr))
	for _, a := range addr {
		if a != nil && SeemsValidAddr(a.Addr) {
			lst = append(lst, a)
		}
	}
	m.Lock()
	defer m.Unlock()
	m.cc = lst
	return m
}

// Bcc sets the (optional) Bcc: email addresses. Last call overrides any previous calls, replacing rather than
// adding to the list.
func (m *Message) Bcc(addr ...*Address) *Message {
	lst := make(addrList, 0, len(addr))
	for _, a := range addr {
		if a != nil && SeemsValidAddr(a.Addr) {
			lst = append(lst, a)
		}
	}
	m.Lock()
	defer m.Unlock()
	m.bcc = lst
	return m
}

// ReplyTo sets the (optional) Reply-To: email address. A `*Address` argument is expected for
// consistency, although only the email address part is used.
func (m *Message) ReplyTo(addr *Address) *Message {
	if addr != nil && !SeemsValidAddr(addr.Addr) {
		addr = nil
	}
	m.Lock()
	defer m.Unlock()
	m.replyTo = addr
	return m
}

// Part adds an alternative part to the message. For a plain-text and/or an HTML body use the
// convenience methods: Text, TextTemplate, Html or HtmlTemplate.
func (m *Message) Part(ctype string, cte CTE, bytes []byte, related ...Related) *Message {
	m.Lock()
	defer m.Unlock()
	m.parts = append(m.parts, &part{
		ctype:   ctype,
		cte:     cte,
		bytes:   bytes,
		related: related,
	})
	m.prepared = false // related may include files
	return m
}

// Text sets the plain-text version of the message body to the provided content.
func (m *Message) Text(text string) *Message {
	m.Lock()
	defer m.Unlock()
	if m.text == nil {
		m.text = &part{}
		m.parts = append(m.parts, m.text)
	}
	*(m.text) = part{
		ctype: "text/plain; charset=utf-8",
		cte:   QuotedPrintable,
		bytes: []byte(text),
	}
	return m
}

// TextTemplate sets the plain-text version of the message body to the provided template.
func (m *Message) TextTemplate(tpl string) *Message {
	var (
		t   *ttpl.Template
		err error
	)
	if tpl != "" {
		t, err = ttpl.New("").Parse(tpl)
		if err != nil {
			m.errors = append(m.errors, errors.New("invalid text template:\n"+tpl+"\nerror: "+err.Error()))
			return m
		}
	}
	m.Lock()
	defer m.Unlock()
	if m.text == nil {
		m.text = &part{}
		m.parts = append(m.parts, m.text)
	}
	*(m.text) = part{
		ctype:  "text/plain; charset=utf-8",
		cte:    QuotedPrintable,
		tplSrc: tpl,
		tpl:    t,
	}
	return m
}

// Html sets the HTML version of the message body to the provided content.
// Optionally, related objects can be specified for inclusion.
func (m *Message) Html(html string, related ...Related) *Message {
	m.Lock()
	defer m.Unlock()
	if m.html == nil {
		m.html = &part{}
		m.parts = append(m.parts, m.html)
	}
	*(m.html) = part{
		ctype:   "text/html; charset=utf-8",
		cte:     QuotedPrintable,
		bytes:   []byte(html),
		related: related,
	}
	m.prepared = false // related may include files
	return m
}

// HtmlTemplate sets the HTML version of the message body to the provided template.
// Optionally, related objects can be specified for inclusion.
func (m *Message) HtmlTemplate(tpl string, related ...Related) *Message {
	var (
		t   *htpl.Template
		err error
	)
	if tpl != "" {
		t, err = htpl.New("").Parse(tpl)
		if err != nil {
			m.errors = append(m.errors, errors.New("invalid html template:\n"+tpl+"\nerror: "+err.Error()))
			return m
		}
	}
	m.Lock()
	defer m.Unlock()
	if m.html == nil {
		m.html = &part{}
		m.parts = append(m.parts, m.html)
	}
	*(m.html) = part{
		ctype:      "text/html; charset=utf-8",
		cte:        QuotedPrintable,
		htmlTplSrc: tpl,
		htmlTpl:    t,
		related:    related,
	}
	m.prepared = false // related may include files
	return m
}

// Attach attaches the files provided as filesystem paths.
func (m *Message) Attach(file ...string) *Message {
	m.Lock()
	defer m.Unlock()
	for _, fileName := range file {
		m.attachments = append(m.attachments, &attachment{fileName: fileName})
	}
	m.prepared = false
	return m
}

// AttachFile attaches a file specified by its filesystem path, setting its name and type
// to the provided values.
func (m *Message) AttachFile(name, ctype, file string) *Message {
	m.Lock()
	defer m.Unlock()
	m.attachments = append(m.attachments, &attachment{
		name:     name,
		ctype:    ctype,
		fileName: file,
	})
	m.prepared = false
	return m
}

// AttachObject creates an attachment with the name, type and data provided.
func (m *Message) AttachObject(name, ctype string, data []byte) *Message {
	m.Lock()
	defer m.Unlock()
	m.attachments = append(m.attachments, &attachment{
		name:  name,
		ctype: ctype,
		data:  data,
	})
	return m
}

func (m *Message) prepare(force bool) {
	if m.prepared && !force {
		return
	}
	allOk := true
	for _, p := range m.parts {
		for _, r := range p.related {
			if r.fileName != "" && (force || len(r.data) == 0) {
				if file, err := ioutil.ReadFile(r.fileName); err == nil {
					r.data = file
				} else {
					m.errors = append(m.errors, errors.New("cannot read file: "+r.fileName+": "+err.Error()))
					allOk = false
				}
			}
		}
	}
	for _, a := range m.attachments {
		if a.fileName != "" && (force || len(a.data) == 0) {
			if file, err := ioutil.ReadFile(a.fileName); err == nil {
				a.data = file
				if a.name == "" {
					a.name = filepath.Base(a.fileName)
				}
				if a.ctype == "" {
					a.ctype = mime.TypeByExtension(filepath.Ext(a.fileName))
				}
			} else {
				m.errors = append(m.errors, errors.New("cannot read file: "+a.fileName+": "+err.Error()))
				allOk = false
			}
		}
	}
	m.prepared = allOk
}

// Prepare reads all the files referenced by the message at attachments or related items.
//
// If the message was already prepared and no new files have been added, it is no-op.
func (m *Message) Prepare() *Message {
	m.Lock()
	defer m.Unlock()
	m.prepare(false)
	return m
}

// PrepareFresh forces a new preparation of the message, even if there were no changes to the referred
// files since the previous one.
func (m *Message) PrepareFresh() *Message {
	m.Lock()
	defer m.Unlock()
	m.prepare(true)
	return m
}

// Compose merges the `data` into the receiver's templates and creates the body of the SMTP message
// to be sent.
func (m *Message) Compose(data interface{}) []byte {
	m.Lock()
	defer m.Unlock()
	var (
		from   *Address
		recpts []*Address
		buf    bytes.Buffer
	)
	switch {
	case m.from != nil:
		from = m.from
	case m.sender != nil && m.sender.address != nil:
		from = m.sender.address
	case defaultSender != nil && defaultSender.address != nil:
		from = defaultSender.address
	}
	if from == nil {
		m.errors = append(m.errors, errors.New("no From address"))
		return []byte{}
	}
	if m.subjectTpl != nil {
		buf.Reset()
		if err := m.subjectTpl.Execute(&buf, data); err != nil {
			m.errors = append(m.errors, errors.New("failed Execute on subject template: "+err.Error()))
		}
		m.subject = make([]byte, buf.Len())
		copy(m.subject, buf.Bytes())
	}
	for partNo, partData := range m.parts {
		switch {
		case partData.tpl != nil:
			buf.Reset()
			if err := partData.tpl.Execute(&buf, data); err != nil {
				m.errors = append(m.errors, errors.New("failed Execute on part["+strconv.Itoa(partNo)+"] template: "+err.Error()))
			}
			partData.bytes = make([]byte, buf.Len())
			copy(partData.bytes, buf.Bytes())
		case partData.htmlTpl != nil:
			buf.Reset()
			if err := partData.htmlTpl.Execute(&buf, data); err != nil {
				m.errors = append(m.errors, errors.New("failed Execute on part["+strconv.Itoa(partNo)+"] html template: "+err.Error()))
			}
			partData.bytes = make([]byte, buf.Len())
			copy(partData.bytes, buf.Bytes())
		}
	}
	if len(m.parts) == 0 {
		m.errors = append(m.errors, errors.New("message has no parts"))
	}
	m.prepare(false)
	if len(m.errors) != 0 {
		return []byte{}
	}

	domain := m.domain
	if len(domain) == 0 {
		domain = []byte(from.Domain())
	}

	ts := []byte(now().In(time.UTC).Format(time.RFC1123Z))
	// hash := sha256.New()
	// hash.Write(ts)
	// hash.Write(m.subject)
	hash := sha256.Sum256(append(ts, m.subject...))
	// uid := Base64Encode(hash.Sum(nil))[:43] // discard padding '='
	uid := Base64Encode(hash[:])[:43] // discard padding '='
	// fmt.Println(string(ts), string(m.subject), ":", string(uid), "--", base64.RawStdEncoding.EncodeToString(hash[:]))

	msg := newBuffer(4096)
	msg.Write("Message-ID: <", uid, '@', domain, ">\r\n")
	msg.Write("Date: ", ts, "\r\n")
	msg.Write("Subject: ", QEncodeIfNeeded(m.subject, 9), "\r\n")
	addr, _ := from.encode(6)
	msg.Write("From: ", addr, "\r\n")
	if m.replyTo != nil && m.replyTo.Addr != "" && m.replyTo.Addr != from.Addr {
		addr, _ = m.replyTo.encode(10)
		msg.Write("Reply-To: ", addr, "\r\n")
	}

	listAddrs := func(list []*Address, offset int) []byte {
		addrs := newBuffer(1024)
		for i, item := range list {
			if i > 0 {
				switch {
				case offset < 75:
					addrs.Write(", ")
					offset += 2
				case offset < 76:
					addrs.Write(",\r\n ")
					offset = 1
				default:
					addrs.Write("\r\n , ")
					offset = 3
				}
			}
			addr, offset = item.encode(offset)
			addrs.Write(addr)
		}
		return addrs.Bytes()
	}

	recpts = m.to
	if len(recpts) == 0 {
		recpts = []*Address{from}
	}
	msg.Write("To: ", listAddrs(recpts, 4), "\r\n")
	if len(m.cc) > 0 {
		msg.Write("Cc: ", listAddrs(m.cc, 4), "\r\n")
	}

	// Do not add BCC addresses into the message - they will show up at all recipients!

	msg.Write("MIME-Version: 1.0\r\n")

	if len(m.attachments) > 0 {
		msg.Write("Content-Type: multipart/mixed;\r\n\tboundary==_m", uid,
			"\r\n\r\n--=_m", uid, "\r\n")
	}

	alt := m.html != nil || len(m.parts) > 1

	if alt {
		msg.Write("Content-Type: multipart/alternative;\r\n\tboundary==_a", uid, "\r\n")
	}

	if m.html != nil && m.text == nil {
		if alt {
			msg.Write("\r\n--=_a", uid, "\r\n")
		}
		msg.Write("Content-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n",
			QuotedPrintableEncode([]byte(htmlToText(string(m.html.bytes)))), "\r\n")
	}
	for partNo, partData := range m.parts {
		if alt {
			msg.Write("\r\n--=_a", uid, "\r\n")
		}
		pn := strconv.Itoa(partNo)
		if len(partData.related) > 0 {
			msg.Write("Content-Type: multipart/related;\r\n\tboundary==_r", pn, uid,
				"\r\n\r\n--=_r", pn, uid, "\r\n")
			// ToDo: substitute the related Ids in content
		}
		switch partData.cte {
		case Base64:
			msg.Write("Content-Type: ", partData.ctype, "\r\nContent-Transfer-Encoding: base64\r\n\r\n",
				Base64Encode(partData.bytes), "\r\n")
		default:
			fallthrough
		case QuotedPrintable:
			msg.Write("Content-Type: ", partData.ctype, "\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n",
				QuotedPrintableEncode(partData.bytes), "\r\n")
		}
		for _, relData := range partData.related {
			msg.Write("\r\n--=_r", pn, uid, "\r\n")
			msg.Write("Content-Type: ", relData.ctype, "\r\nContent-Transfer-Encoding: base64\r\n\r\n",
				Base64Encode(relData.data), "\r\n")
		}
		if len(partData.related) > 0 {
			msg.Write("\r\n--=_r", pn, uid, "--\r\n")
		}
	}
	if alt {
		msg.Write("\r\n--=_a", uid, "--\r\n")
	}

	for _, attData := range m.attachments {
		msg.Write("\r\n--=_m", uid, "\r\n")
		msg.Write("Content-Type: ", attData.ctype,
			"\r\nContent-Disposition: attachment;\r\n\tfilename=", attData.name,
			"\r\nContent-Transfer-Encoding: base64\r\n\r\n",
			Base64Encode(attData.data), "\r\n")
	}

	if len(m.attachments) > 0 {
		msg.Write("\r\n--=_m", uid, "--\r\n")
	}

	return msg.Bytes()
}

// FromAddr returns the email address that the message would be sent from.
func (m *Message) FromAddr() string {
	m.RLock()
	defer m.RUnlock()
	var from *Address
	switch {
	case m.from != nil:
		from = m.from
	case m.sender != nil && m.sender.address != nil:
		from = m.sender.address
	case defaultSender != nil && defaultSender.address != nil:
		from = defaultSender.address
	}
	if from != nil {
		return from.Addr
	}
	return ""
}

// RecipientAddrs returns a list of email addresses with all the recipients for the message.
//
// It includes addresses from the To, CC and BCC fields.
func (m *Message) RecipientAddrs() []string {
	m.RLock()
	defer m.RUnlock()
	to := make([]string, 0, len(m.to)+len(m.cc)+len(m.bcc)+1)
	seen := map[string]struct{}{}
	if len(m.to) == 0 {
		addr := m.FromAddr()
		to = append(to, addr)
		seen[addr] = struct{}{}
	}
	for _, val := range m.to {
		addr := val.Addr
		if _, s := seen[addr]; !s {
			to = append(to, addr)
			seen[addr] = struct{}{}
		}
	}
	for _, val := range m.cc {
		addr := val.Addr
		if _, s := seen[addr]; !s {
			to = append(to, addr)
			seen[addr] = struct{}{}
		}
	}
	for _, val := range m.bcc {
		addr := val.Addr
		if _, s := seen[addr]; !s {
			to = append(to, addr)
			seen[addr] = struct{}{}
		}
	}
	return to
}

// HasErrors checks if there are any errors associated with the receiver
func (m *Message) HasErrors() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.errors) > 0
}

// Errors returns the list of errors associated with the receiver, then resets the internal list.
func (m *Message) Errors() (errs []error) {
	m.Lock()
	defer m.Unlock()
	errs, m.errors = m.errors, nil
	return
}

// NewMessage creates a new Message, deep-copying from `msg`, if provided.
func NewMessage(msg *Message) *Message {
	if msg == nil {
		return &Message{prepared: true}
	}
	msg.RLock()
	defer msg.RUnlock()
	m := &Message{
		domain:        msg.domain,
		sender:        msg.sender,
		subject:       msg.subject,
		subjectTplSrc: msg.subjectTplSrc,
		from:          msg.from.Clone(),
		replyTo:       msg.replyTo.Clone(),
		to:            msg.to.Clone(),
		cc:            msg.cc.Clone(),
		bcc:           msg.bcc.Clone(),
		prepared:      msg.prepared,
	}
	if msg.subjectTplSrc != "" {
		// the template source was already parsed successfully once, so it is guaranteed to be valid
		m.subjectTpl, _ = ttpl.New("").Parse(msg.subjectTplSrc)
	}
	m.parts = make([]*part, len(msg.parts))
	for i, partData := range msg.parts {
		p := &part{
			ctype:      partData.ctype,
			cte:        partData.cte,
			tplSrc:     partData.tplSrc,
			htmlTplSrc: partData.htmlTplSrc,
			// related    []Related
		}
		if len(partData.bytes) > 0 {
			p.bytes = make([]byte, len(partData.bytes))
			copy(p.bytes, partData.bytes)
		}
		if partData.tplSrc != "" {
			// the template source was already parsed successfully once, so it is guaranteed to be valid
			p.tpl, _ = ttpl.New("").Parse(partData.tplSrc)
		}
		if partData.htmlTplSrc != "" {
			// the template source was already parsed successfully once, so it is guaranteed to be valid
			p.htmlTpl, _ = htpl.New("").Parse(partData.htmlTplSrc)
		}
		if len(partData.related) > 0 {
			p.related = make([]Related, len(partData.related))
			copy(p.related, partData.related)
			// do not copy partData.related.data, to save memory; it is never updated in place
		}
		if msg.text == partData {
			m.text = p
		}
		if msg.html == partData {
			m.html = p
		}
		m.parts[i] = p
	}
	m.attachments = make([]*attachment, len(msg.attachments))
	for i, attData := range msg.attachments {
		m.attachments[i] = attData
		// do not copy attData.data, to save memory; it is never updated in place
	}
	return m
}

// QuickMessage creates a Message with the subject and the body provided. Alternative text and HTML
// body versions can be provided, in this order.
func QuickMessage(subject string, body ...string) *Message {
	msg := &Message{subject: []byte(subject), prepared: true}
	if len(body) > 0 {
		msg.Text(body[0])
	}
	if len(body) > 1 {
		msg.Html(body[1])
	}
	return msg
}

type part struct {
	ctype      string
	cte        CTE
	bytes      []byte
	tplSrc     string
	tpl        *ttpl.Template
	htmlTplSrc string
	htmlTpl    *htpl.Template
	related    []Related
}

// Related represents a multipart/related item.
type Related struct {
	id       string
	ctype    string
	fileName string
	data     []byte
}

// RelatedFile creates a Related structure from the provided file information.
func RelatedFile(id, ctype, file string) Related {
	return Related{
		id:       id,
		ctype:    ctype,
		fileName: file,
	}
}

// RelatedObject creates a Related structure from the provided data.
func RelatedObject(id, ctype string, data []byte) Related {
	return Related{
		id:    id,
		ctype: ctype,
		data:  data,
	}
}

type attachment struct {
	name     string
	ctype    string
	fileName string
	data     []byte
}
