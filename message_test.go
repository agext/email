package email

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agext/uuid"
)

type messageObj struct {
	name, ctype string
	cte         CTE
	bytes       []byte
	related     []Related
}
type messageIn struct {
	domain        string
	subject       string
	subjectTpl    string
	sender        *Sender
	from, replyTo *Address
	to, cc, bcc   []*Address
	parts         []messageObj
	text, textTpl string
	html, htmlTpl string
	rel           []Related
	attachments   []messageObj
}

type messageTestCase struct {
	src    messageIn
	data   interface{}
	date   time.Time
	expOut []byte
	expErr []string
}

func forceNow(unix int64) {
	now = func() time.Time { return time.Unix(unix, 0) }
}

func Test_Compose(t *testing.T) {
	date := time.Date(2013, 8, 30, 9, 10, 11, 0, time.UTC)
	workDir, _ := os.Getwd()
	uid := []byte(uuid.New().Hex())
	newUUID = func() []byte { return uid }
	cases := []messageTestCase{
		{
			src: messageIn{
				subject: "Test #1",
				from:    &Address{"test name", "test@example.com"},
				text:    "Short test message",
			},
			expOut: []byte("Message-ID: <" + string(uid) + "@example.com>\r\n" +
				"Date: Fri, 30 Aug 2013 09:10:11 +0000\r\n" +
				"Subject: Test #1\r\n" +
				"From: \"test name\" <test@example.com>\r\n" +
				"To: \"test name\" <test@example.com>\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"Short test message\r\n"),
		},
		{
			src: messageIn{
				subject: "Test #2",
				from:    &Address{"accented nåmé", "test@example.com"},
				html:    "<head><style>.test {color:red}</style></head><body>Html <b>test</b> message</body>",
			},
			expOut: []byte("Message-ID: <" + string(uid) + "@example.com>\r\n" +
				"Date: Fri, 30 Aug 2013 09:10:11 +0000\r\n" +
				"Subject: Test #2\r\n" +
				"From: =?utf-8?q?accented_n=C3=A5m=C3=A9?= <test@example.com>\r\n" +
				"To: =?utf-8?q?accented_n=C3=A5m=C3=A9?= <test@example.com>\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/alternative;\r\n" +
				"\tboundary=B_a_" + string(uid) + "\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"Html test message\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/html; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"<head><style>.test {color:red}</style></head><body>Html <b>test</b> message=\r\n" +
				"</body>\r\n\r\n" +
				"--B_a_" + string(uid) + "--\r\n"),
		},
		{
			src: messageIn{
				subject: "Test… #3",
				from:    &Address{"accented nåmé", "test@example.com"},
				to: []*Address{{"Δεσωρε αππελλανθυρ υθ μει, ", "test1@example.com"},
					{"αν ηαβεο ομνες νυμκυαμ μεα.", "test2@example.com"}},
				html: "<head><style>.test {color:red}</style></head>\r\n<body>Html <b>test</b> message = Αδ φιξ αλικυιπ ινφιδυντ, ηις εξ σαπερεθ δετρασθο σαεφολα, αδ δολορ αλικυανδο ηας.</body>",
				attachments: []messageObj{
					{name: "test-file.txt", ctype: "text/plain", bytes: []byte("Δεσωρε αππελλανθυρ υθ μει, αν ηαβεο ομνες νυμκυαμ μεα. Αδ φιξ αλικυιπ ινφιδυντ, ηις εξ σαπερεθ δετρασθο σαεφολα, αδ δολορ αλικυανδο ηας. Ευ πυρθο ιυδισο εως, φισι σωνσεκυαθ πρι ευ. Ασυμ σοντεντιωνες ιυς ει, ει κυαεκυε ινσωλενς σενσιβυς κυο. Εξ κυωτ αλιενυμ ηις, συ πρω σονσυλατυ μεδιοσριθαθεμ. Τιβικυε ινστρυσθιορ κυι νο, ευμ ιδ κυοδσι τασιμαθες αδωλεσενς.")},
				},
			},
			expOut: []byte("Message-ID: <" + string(uid) + "@example.com>\r\n" +
				"Date: Fri, 30 Aug 2013 09:10:11 +0000\r\n" +
				"Subject: =?utf-8?q?Test=E2=80=A6_#3?=\r\n" +
				"From: =?utf-8?q?accented_n=C3=A5m=C3=A9?= <test@example.com>\r\n" +
				"To: =?utf-8?q?=CE=94=CE=B5=CF=83=CF=89=CF=81=CE=B5_=CE=B1=CF=80=CF=80?=\r\n" +
				" =?utf-8?q?=CE=B5=CE=BB=CE=BB=CE=B1=CE=BD=CE=B8=CF=85=CF=81_=CF=85=CE=B8_?=\r\n" +
				" =?utf-8?q?=CE=BC=CE=B5=CE=B9,_?= <test1@example.com>, =?utf-8?q?=CE=B1?=\r\n" +
				" =?utf-8?q?=CE=BD_=CE=B7=CE=B1=CE=B2=CE=B5=CE=BF_=CE=BF=CE=BC=CE=BD=CE=B5?=\r\n" +
				" =?utf-8?q?=CF=82_=CE=BD=CF=85=CE=BC=CE=BA=CF=85=CE=B1=CE=BC_=CE=BC=CE=B5?=\r\n" +
				" =?utf-8?q?=CE=B1.?= <test2@example.com>\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/mixed;\r\n" +
				"\tboundary=B_m_" + string(uid) + "\r\n\r\n" +
				"--B_m_" + string(uid) + "\r\n" +
				"Content-Type: multipart/alternative;\r\n" +
				"\tboundary=B_a_" + string(uid) + "\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"Html test message =3D =CE=91=CE=B4 =CF=86=CE=B9=CE=BE =CE=B1=CE=BB=CE=B9=\r\n" +
				"=CE=BA=CF=85=CE=B9=CF=80 =CE=B9=CE=BD=CF=86=CE=B9=CE=B4=CF=85=CE=BD=CF=84, =\r\n" +
				"=CE=B7=CE=B9=CF=82 =CE=B5=CE=BE =CF=83=CE=B1=CF=80=CE=B5=CF=81=CE=B5=CE=B8 =\r\n" +
				"=CE=B4=CE=B5=CF=84=CF=81=CE=B1=CF=83=CE=B8=CE=BF =CF=83=CE=B1=CE=B5=CF=86=\r\n" +
				"=CE=BF=CE=BB=CE=B1, =CE=B1=CE=B4 =CE=B4=CE=BF=CE=BB=CE=BF=CF=81 =CE=B1=\r\n" +
				"=CE=BB=CE=B9=CE=BA=CF=85=CE=B1=CE=BD=CE=B4=CE=BF =CE=B7=CE=B1=CF=82.\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/html; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"<head><style>.test {color:red}</style></head>=0D=0A<body>Html <b>test</b> m=\r\n" +
				"essage =3D =CE=91=CE=B4 =CF=86=CE=B9=CE=BE =CE=B1=CE=BB=CE=B9=CE=BA=CF=85=\r\n" +
				"=CE=B9=CF=80 =CE=B9=CE=BD=CF=86=CE=B9=CE=B4=CF=85=CE=BD=CF=84, =CE=B7=CE=B9=\r\n" +
				"=CF=82 =CE=B5=CE=BE =CF=83=CE=B1=CF=80=CE=B5=CF=81=CE=B5=CE=B8 =CE=B4=CE=B5=\r\n" +
				"=CF=84=CF=81=CE=B1=CF=83=CE=B8=CE=BF =CF=83=CE=B1=CE=B5=CF=86=CE=BF=CE=BB=\r\n" +
				"=CE=B1, =CE=B1=CE=B4 =CE=B4=CE=BF=CE=BB=CE=BF=CF=81 =CE=B1=CE=BB=CE=B9=\r\n" +
				"=CE=BA=CF=85=CE=B1=CE=BD=CE=B4=CE=BF =CE=B7=CE=B1=CF=82.</body>\r\n\r\n" +
				"--B_a_" + string(uid) + "--\r\n\r\n" +
				"--B_m_" + string(uid) + "\r\n" +
				"Content-Type: text/plain\r\n" +
				"Content-Disposition: attachment;\r\n" +
				"\tfilename=\"test-file.txt\"\r\n" +
				"Content-Transfer-Encoding: base64\r\n\r\n" +
				"zpTOtc+Dz4nPgc61IM6xz4DPgM61zrvOu86xzr3OuM+Fz4Egz4XOuCDOvM61zrksIM6xzr0gzrfO\r\n" +
				"sc6yzrXOvyDOv868zr3Otc+CIM69z4XOvM66z4XOsc68IM68zrXOsS4gzpHOtCDPhs65zr4gzrHO\r\n" +
				"u865zrrPhc65z4AgzrnOvc+GzrnOtM+Fzr3PhCwgzrfOuc+CIM61zr4gz4POsc+AzrXPgc61zrgg\r\n" +
				"zrTOtc+Ez4HOsc+DzrjOvyDPg86xzrXPhs6/zrvOsSwgzrHOtCDOtM6/zrvOv8+BIM6xzrvOuc66\r\n" +
				"z4XOsc69zrTOvyDOt86xz4IuIM6Vz4Ugz4DPhc+BzrjOvyDOuc+FzrTOuc+Dzr8gzrXPic+CLCDP\r\n" +
				"hs65z4POuSDPg8+Jzr3Pg861zrrPhc6xzrggz4DPgc65IM61z4UuIM6Rz4PPhc68IM+Dzr/Ovc+E\r\n" +
				"zrXOvc+EzrnPic69zrXPgiDOuc+Fz4IgzrXOuSwgzrXOuSDOus+FzrHOtc66z4XOtSDOuc69z4PP\r\n" +
				"ic67zrXOvc+CIM+DzrXOvc+DzrnOss+Fz4IgzrrPhc6/LiDOlc6+IM66z4XPic+EIM6xzrvOuc61\r\n" +
				"zr3Phc68IM63zrnPgiwgz4PPhSDPgM+Bz4kgz4POv869z4PPhc67zrHPhM+FIM68zrXOtM65zr/P\r\n" +
				"g8+BzrnOuM6xzrjOtc68LiDOpM65zrLOuc66z4XOtSDOuc69z4PPhM+Bz4XPg864zrnOv8+BIM66\r\n" +
				"z4XOuSDOvc6/LCDOtc+FzrwgzrnOtCDOus+Fzr/OtM+Dzrkgz4TOsc+DzrnOvM6xzrjOtc+CIM6x\r\n" +
				"zrTPic67zrXPg861zr3Pgi4=\r\n\r\n" +
				"--B_m_" + string(uid) + "--\r\n"),
		},
		{
			src: messageIn{
				subject: "Test… #4",
				from:    &Address{"accented nåmé", "test@example.com"},
				to: []*Address{{"Δεσωρε αππελλανθυρ υθ μει, ", "test1@example.com"},
					{"αν ηαβεο ομνες νυμκυαμ μεα.", "test2@example.com"}},
				replyTo: &Address{"different name", "test-reply@example.com"},
				html:    "<head><style>.test {color:red}</style></head>\r\n<body>Html <b>test</b> message = Αδ φιξ αλικυιπ ινφιδυντ, ηις εξ σαπερεθ δετρασθο σαεφολα, αδ δολορ αλικυανδο ηας.</body>",
				attachments: []messageObj{
					{name: filepath.Join(workDir, "test-file.txt")},
				},
			},
			expOut: []byte("Message-ID: <" + string(uid) + "@example.com>\r\n" +
				"Date: Fri, 30 Aug 2013 09:10:11 +0000\r\n" +
				"Subject: =?utf-8?q?Test=E2=80=A6_#4?=\r\n" +
				"From: =?utf-8?q?accented_n=C3=A5m=C3=A9?= <test@example.com>\r\n" +
				"Reply-To: \"different name\" <test-reply@example.com>\r\n" +
				"To: =?utf-8?q?=CE=94=CE=B5=CF=83=CF=89=CF=81=CE=B5_=CE=B1=CF=80=CF=80?=\r\n" +
				" =?utf-8?q?=CE=B5=CE=BB=CE=BB=CE=B1=CE=BD=CE=B8=CF=85=CF=81_=CF=85=CE=B8_?=\r\n" +
				" =?utf-8?q?=CE=BC=CE=B5=CE=B9,_?= <test1@example.com>, =?utf-8?q?=CE=B1?=\r\n" +
				" =?utf-8?q?=CE=BD_=CE=B7=CE=B1=CE=B2=CE=B5=CE=BF_=CE=BF=CE=BC=CE=BD=CE=B5?=\r\n" +
				" =?utf-8?q?=CF=82_=CE=BD=CF=85=CE=BC=CE=BA=CF=85=CE=B1=CE=BC_=CE=BC=CE=B5?=\r\n" +
				" =?utf-8?q?=CE=B1.?= <test2@example.com>\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/mixed;\r\n" +
				"\tboundary=B_m_" + string(uid) + "\r\n\r\n" +
				"--B_m_" + string(uid) + "\r\n" +
				"Content-Type: multipart/alternative;\r\n" +
				"\tboundary=B_a_" + string(uid) + "\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"Html test message =3D =CE=91=CE=B4 =CF=86=CE=B9=CE=BE =CE=B1=CE=BB=CE=B9=\r\n" +
				"=CE=BA=CF=85=CE=B9=CF=80 =CE=B9=CE=BD=CF=86=CE=B9=CE=B4=CF=85=CE=BD=CF=84, =\r\n" +
				"=CE=B7=CE=B9=CF=82 =CE=B5=CE=BE =CF=83=CE=B1=CF=80=CE=B5=CF=81=CE=B5=CE=B8 =\r\n" +
				"=CE=B4=CE=B5=CF=84=CF=81=CE=B1=CF=83=CE=B8=CE=BF =CF=83=CE=B1=CE=B5=CF=86=\r\n" +
				"=CE=BF=CE=BB=CE=B1, =CE=B1=CE=B4 =CE=B4=CE=BF=CE=BB=CE=BF=CF=81 =CE=B1=\r\n" +
				"=CE=BB=CE=B9=CE=BA=CF=85=CE=B1=CE=BD=CE=B4=CE=BF =CE=B7=CE=B1=CF=82.\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/html; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"<head><style>.test {color:red}</style></head>=0D=0A<body>Html <b>test</b> m=\r\n" +
				"essage =3D =CE=91=CE=B4 =CF=86=CE=B9=CE=BE =CE=B1=CE=BB=CE=B9=CE=BA=CF=85=\r\n" +
				"=CE=B9=CF=80 =CE=B9=CE=BD=CF=86=CE=B9=CE=B4=CF=85=CE=BD=CF=84, =CE=B7=CE=B9=\r\n" +
				"=CF=82 =CE=B5=CE=BE =CF=83=CE=B1=CF=80=CE=B5=CF=81=CE=B5=CE=B8 =CE=B4=CE=B5=\r\n" +
				"=CF=84=CF=81=CE=B1=CF=83=CE=B8=CE=BF =CF=83=CE=B1=CE=B5=CF=86=CE=BF=CE=BB=\r\n" +
				"=CE=B1, =CE=B1=CE=B4 =CE=B4=CE=BF=CE=BB=CE=BF=CF=81 =CE=B1=CE=BB=CE=B9=\r\n" +
				"=CE=BA=CF=85=CE=B1=CE=BD=CE=B4=CE=BF =CE=B7=CE=B1=CF=82.</body>\r\n\r\n" +
				"--B_a_" + string(uid) + "--\r\n\r\n" +
				"--B_m_" + string(uid) + "\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Content-Disposition: attachment;\r\n" +
				"\tfilename=\"test-file.txt\"\r\n" +
				"Content-Transfer-Encoding: base64\r\n\r\n" +
				"zpTOtc+Dz4nPgc61IM6xz4DPgM61zrvOu86xzr3OuM+Fz4Egz4XOuCDOvM61zrksIM6xzr0gzrfO\r\n" +
				"sc6yzrXOvyDOv868zr3Otc+CIM69z4XOvM66z4XOsc68IM68zrXOsS4gzpHOtCDPhs65zr4gzrHO\r\n" +
				"u865zrrPhc65z4AgzrnOvc+GzrnOtM+Fzr3PhCwgzrfOuc+CIM61zr4gz4POsc+AzrXPgc61zrgg\r\n" +
				"zrTOtc+Ez4HOsc+DzrjOvyDPg86xzrXPhs6/zrvOsSwgzrHOtCDOtM6/zrvOv8+BIM6xzrvOuc66\r\n" +
				"z4XOsc69zrTOvyDOt86xz4IuIM6Vz4Ugz4DPhc+BzrjOvyDOuc+FzrTOuc+Dzr8gzrXPic+CLCDP\r\n" +
				"hs65z4POuSDPg8+Jzr3Pg861zrrPhc6xzrggz4DPgc65IM61z4UuIM6Rz4PPhc68IM+Dzr/Ovc+E\r\n" +
				"zrXOvc+EzrnPic69zrXPgiDOuc+Fz4IgzrXOuSwgzrXOuSDOus+FzrHOtc66z4XOtSDOuc69z4PP\r\n" +
				"ic67zrXOvc+CIM+DzrXOvc+DzrnOss+Fz4IgzrrPhc6/LiDOlc6+IM66z4XPic+EIM6xzrvOuc61\r\n" +
				"zr3Phc68IM63zrnPgiwgz4PPhSDPgM+Bz4kgz4POv869z4PPhc67zrHPhM+FIM68zrXOtM65zr/P\r\n" +
				"g8+BzrnOuM6xzrjOtc68LiDOpM65zrLOuc66z4XOtSDOuc69z4PPhM+Bz4XPg864zrnOv8+BIM66\r\n" +
				"z4XOuSDOvc6/LCDOtc+FzrwgzrnOtCDOus+Fzr/OtM+Dzrkgz4TOsc+DzrnOvM6xzrjOtc+CIM6x\r\n" +
				"zrTPic67zrXPg861zr3Pgi4K\r\n\r\n" +
				"--B_m_" + string(uid) + "--\r\n"),
		},
		{
			src: messageIn{
				subjectTpl: "Test {{.name}}",
				from:       &Address{"test name", "test@example.com"},
				textTpl:    "Hi {{.name}}!",
				htmlTpl:    "<head></head><body>Hi {{.name}}!</body>",
			},
			data: map[string]string{"name": "John & Jill"},
			expOut: []byte("Message-ID: <" + string(uid) + "@example.com>\r\n" +
				"Date: Fri, 30 Aug 2013 09:10:11 +0000\r\n" +
				"Subject: Test John & Jill\r\n" +
				"From: \"test name\" <test@example.com>\r\n" +
				"To: \"test name\" <test@example.com>\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/alternative;\r\n" +
				"\tboundary=B_a_" + string(uid) + "\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"Hi John & Jill!\r\n\r\n" +
				"--B_a_" + string(uid) + "\r\n" +
				"Content-Type: text/html; charset=utf-8\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
				"<head></head><body>Hi John &amp; Jill!</body>\r\n\r\n" +
				"--B_a_" + string(uid) + "--\r\n"),
		},
	}

	for i, c := range cases {
		msg := NewMessage(nil).Domain(c.src.domain).Subject(c.src.subject).
			setSender(c.src.sender).From(c.src.from).ReplyTo(c.src.replyTo).
			To(c.src.to...).Cc(c.src.cc...).Bcc(c.src.bcc...)
		if c.src.subjectTpl != "" {
			msg.SubjectTemplate(c.src.subjectTpl)
		}
		if c.src.text != "" {
			msg.Text(c.src.text)
		}
		if c.src.textTpl != "" {
			msg.TextTemplate(c.src.textTpl)
		}
		if c.src.html != "" {
			msg.Html(c.src.html, c.src.rel...)
		}
		if c.src.htmlTpl != "" {
			msg.HtmlTemplate(c.src.htmlTpl, c.src.rel...)
		}
		for _, partData := range c.src.parts {
			msg.Part(partData.ctype, partData.cte, partData.bytes, partData.related...)
		}
		for _, attData := range c.src.attachments {
			if len(attData.bytes) > 0 {
				msg.AttachObject(attData.name, attData.ctype, attData.bytes)
			} else {
				msg.Attach(attData.name)
			}
		}
		if !c.date.IsZero() {
			forceNow(c.date.Unix())
		} else {
			forceNow(date.Unix())
		}
		act := msg.Compose(c.data)
		if !bytes.Equal(act, c.expOut) {
			t.Errorf("(*Message).Compose [%d]: got (len=%d)\n%s\nwant (len=%d)\n%s", i, len(act), act, len(c.expOut), c.expOut)
		}
		if len(msg.errors) != len(c.expErr) {
			t.Errorf("(*Message).Compose [%d]: got %d errors, want %d:\n", i, len(msg.errors), len(c.expErr))
			for _, err := range msg.errors {
				t.Errorf("%s\n", err.Error())
			}
			t.Error("was expecting:\n")
			for _, err := range c.expErr {
				t.Errorf("%s\n", err)
			}
		}
	}
}
