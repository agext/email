package email

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

type encodingTestCase struct {
	src, exp []byte
}

func Test_Base64Encode(t *testing.T) {
	for i := 1; i < 1024; i++ {
		src := make([]byte, i)
		rand.Read(src)
		el := base64.StdEncoding.EncodedLen(i)
		b64 := make([]byte, el)
		base64.StdEncoding.Encode(b64, src)
		exp := make([]byte, 0, el+el/76*2)
		for len(b64) > 76 {
			exp = append(exp, b64[:76]...)
			exp = append(exp, '\r', '\n')
			b64 = b64[76:]
		}
		exp = append(exp, b64...)
		if act := Base64Encode(src); !bytes.Equal(act, exp) {
			t.Errorf("Base64Encode(%d): got (len=%d)\n%s\nwant (len=%d)\n%s", i, len(act), act, len(exp), exp)
		}
	}
}

func Test_QuotedPrintableEncode(t *testing.T) {
	cases := []encodingTestCase{
		{[]byte("test "), []byte("test =")},
		{[]byte("test\n me"), []byte("test=0A me")},
		{[]byte("test\\/me=again…"), []byte("test\\/me=3Dagain=E2=80=A6")},
		{[]byte("Lorem ipsum dolor sit amet, no sit enim fugit, solum omittam evertitur qui cu. Usu ad sonet facilisis, cu partem platonem conceptam has. Tincidunt scribentur nec ex, eu hinc quodsi consequat quo, ex est labore fuisset. Vel semper salutatus ne."),
			[]byte("Lorem ipsum dolor sit amet, no sit enim fugit, solum omittam evertitur qui =\r\n" +
				"cu. Usu ad sonet facilisis, cu partem platonem conceptam has. Tincidunt scr=\r\n" +
				"ibentur nec ex, eu hinc quodsi consequat quo, ex est labore fuisset. Vel se=\r\n" +
				"mper salutatus ne.")},
		{[]byte("Δεσωρε αππελλανθυρ υθ μει, αν ηαβεο ομνες νυμκυαμ μεα. Αδ φιξ αλικυιπ ινφιδυντ, ηις εξ σαπερεθ δετρασθο σαεφολα, αδ δολορ αλικυανδο ηας. Ευ πυρθο ιυδισο εως, φισι σωνσεκυαθ πρι ευ. Ασυμ σοντεντιωνες ιυς ει, ει κυαεκυε ινσωλενς σενσιβυς κυο. Εξ κυωτ αλιενυμ ηις, συ πρω σονσυλατυ μεδιοσριθαθεμ. Τιβικυε ινστρυσθιορ κυι νο, ευμ ιδ κυοδσι τασιμαθες αδωλεσενς."),
			[]byte("=CE=94=CE=B5=CF=83=CF=89=CF=81=CE=B5 =CE=B1=CF=80=CF=80=CE=B5=CE=BB=CE=BB=\r\n" +
				"=CE=B1=CE=BD=CE=B8=CF=85=CF=81 =CF=85=CE=B8 =CE=BC=CE=B5=CE=B9, =CE=B1=\r\n" +
				"=CE=BD =CE=B7=CE=B1=CE=B2=CE=B5=CE=BF =CE=BF=CE=BC=CE=BD=CE=B5=CF=82 =CE=BD=\r\n" +
				"=CF=85=CE=BC=CE=BA=CF=85=CE=B1=CE=BC =CE=BC=CE=B5=CE=B1. =CE=91=CE=B4 =\r\n" +
				"=CF=86=CE=B9=CE=BE =CE=B1=CE=BB=CE=B9=CE=BA=CF=85=CE=B9=CF=80 =CE=B9=CE=BD=\r\n" +
				"=CF=86=CE=B9=CE=B4=CF=85=CE=BD=CF=84, =CE=B7=CE=B9=CF=82 =CE=B5=CE=BE =\r\n" +
				"=CF=83=CE=B1=CF=80=CE=B5=CF=81=CE=B5=CE=B8 =CE=B4=CE=B5=CF=84=CF=81=CE=B1=\r\n" +
				"=CF=83=CE=B8=CE=BF =CF=83=CE=B1=CE=B5=CF=86=CE=BF=CE=BB=CE=B1, =CE=B1=CE=B4=\r\n" +
				" =CE=B4=CE=BF=CE=BB=CE=BF=CF=81 =CE=B1=CE=BB=CE=B9=CE=BA=CF=85=CE=B1=CE=BD=\r\n" +
				"=CE=B4=CE=BF =CE=B7=CE=B1=CF=82. =CE=95=CF=85 =CF=80=CF=85=CF=81=CE=B8=\r\n" +
				"=CE=BF =CE=B9=CF=85=CE=B4=CE=B9=CF=83=CE=BF =CE=B5=CF=89=CF=82, =CF=86=\r\n" +
				"=CE=B9=CF=83=CE=B9 =CF=83=CF=89=CE=BD=CF=83=CE=B5=CE=BA=CF=85=CE=B1=CE=B8 =\r\n" +
				"=CF=80=CF=81=CE=B9 =CE=B5=CF=85. =CE=91=CF=83=CF=85=CE=BC =CF=83=CE=BF=\r\n" +
				"=CE=BD=CF=84=CE=B5=CE=BD=CF=84=CE=B9=CF=89=CE=BD=CE=B5=CF=82 =CE=B9=CF=85=\r\n" +
				"=CF=82 =CE=B5=CE=B9, =CE=B5=CE=B9 =CE=BA=CF=85=CE=B1=CE=B5=CE=BA=CF=85=\r\n" +
				"=CE=B5 =CE=B9=CE=BD=CF=83=CF=89=CE=BB=CE=B5=CE=BD=CF=82 =CF=83=CE=B5=CE=BD=\r\n" +
				"=CF=83=CE=B9=CE=B2=CF=85=CF=82 =CE=BA=CF=85=CE=BF. =CE=95=CE=BE =CE=BA=\r\n" +
				"=CF=85=CF=89=CF=84 =CE=B1=CE=BB=CE=B9=CE=B5=CE=BD=CF=85=CE=BC =CE=B7=CE=B9=\r\n" +
				"=CF=82, =CF=83=CF=85 =CF=80=CF=81=CF=89 =CF=83=CE=BF=CE=BD=CF=83=CF=85=\r\n" +
				"=CE=BB=CE=B1=CF=84=CF=85 =CE=BC=CE=B5=CE=B4=CE=B9=CE=BF=CF=83=CF=81=CE=B9=\r\n" +
				"=CE=B8=CE=B1=CE=B8=CE=B5=CE=BC. =CE=A4=CE=B9=CE=B2=CE=B9=CE=BA=CF=85=CE=B5 =\r\n" +
				"=CE=B9=CE=BD=CF=83=CF=84=CF=81=CF=85=CF=83=CE=B8=CE=B9=CE=BF=CF=81 =CE=BA=\r\n" +
				"=CF=85=CE=B9 =CE=BD=CE=BF, =CE=B5=CF=85=CE=BC =CE=B9=CE=B4 =CE=BA=CF=85=\r\n" +
				"=CE=BF=CE=B4=CF=83=CE=B9 =CF=84=CE=B1=CF=83=CE=B9=CE=BC=CE=B1=CE=B8=CE=B5=\r\n" +
				"=CF=82 =CE=B1=CE=B4=CF=89=CE=BB=CE=B5=CF=83=CE=B5=CE=BD=CF=82.")},
	}
	for _, c := range cases {
		if act := QuotedPrintableEncode(c.src); !bytes.Equal(act, c.exp) {
			t.Errorf("QuotedPrintableEncode: got (len=%d)\n%s\nwant (len=%d)\n%s", len(act), act, len(c.exp), c.exp)
		}
	}
}

func Test_QEncode(t *testing.T) {
	cases := []encodingTestCase{
		{[]byte("test "), []byte("=?utf-8?q?test_?=")},
		{[]byte("test\n me"), []byte("=?utf-8?q?test=0A_me?=")},
		{[]byte("test\\/me=again…"), []byte("=?utf-8?q?test\\/me=3Dagain=E2=80=A6?=")},
		{[]byte("Lorem ipsum dolor sit amet, no sit enim fugit, solum omittam evertitur qui cu. Usu ad sonet facilisis, cu partem platonem conceptam has. Tincidunt scribentur nec ex, eu hinc quodsi consequat quo, ex est labore fuisset. Vel semper salutatus ne."),
			[]byte("=?utf-8?q?Lorem_ipsum_dolor_sit_amet,_no_s?=\r\n" +
				" =?utf-8?q?it_enim_fugit,_solum_omittam_evertitur_qui_cu._Usu_ad_sonet_fac?=\r\n" +
				" =?utf-8?q?ilisis,_cu_partem_platonem_conceptam_has._Tincidunt_scribentur_?=\r\n" +
				" =?utf-8?q?nec_ex,_eu_hinc_quodsi_consequat_quo,_ex_est_labore_fuisset._Ve?=\r\n" +
				" =?utf-8?q?l_semper_salutatus_ne.?=")},
		{[]byte("Δεσωρε αππελλανθυρ υθ μει, αν ηαβεο ομνες νυμκυαμ μεα. Αδ φιξ αλικυιπ ινφιδυντ, ηις εξ σαπερεθ δετρασθο σαεφολα, αδ δολορ αλικυανδο ηας. Ευ πυρθο ιυδισο εως, φισι σωνσεκυαθ πρι ευ. Ασυμ σοντεντιωνες ιυς ει, ει κυαεκυε ινσωλενς σενσιβυς κυο. Εξ κυωτ αλιενυμ ηις, συ πρω σονσυλατυ μεδιοσριθαθεμ. Τιβικυε ινστρυσθιορ κυι νο, ευμ ιδ κυοδσι τασιμαθες αδωλεσενς."),
			[]byte("=?utf-8?q?=CE=94=CE=B5=CF=83=CF=89=CF=81?=\r\n" +
				" =?utf-8?q?=CE=B5_=CE=B1=CF=80=CF=80=CE=B5=CE=BB=CE=BB=CE=B1=CE=BD=CE=B8?=\r\n" +
				" =?utf-8?q?=CF=85=CF=81_=CF=85=CE=B8_=CE=BC=CE=B5=CE=B9,_=CE=B1=CE=BD_?=\r\n" +
				" =?utf-8?q?=CE=B7=CE=B1=CE=B2=CE=B5=CE=BF_=CE=BF=CE=BC=CE=BD=CE=B5=CF=82_?=\r\n" +
				" =?utf-8?q?=CE=BD=CF=85=CE=BC=CE=BA=CF=85=CE=B1=CE=BC_=CE=BC=CE=B5=CE=B1._?=\r\n" +
				" =?utf-8?q?=CE=91=CE=B4_=CF=86=CE=B9=CE=BE_=CE=B1=CE=BB=CE=B9=CE=BA=CF=85?=\r\n" +
				" =?utf-8?q?=CE=B9=CF=80_=CE=B9=CE=BD=CF=86=CE=B9=CE=B4=CF=85=CE=BD=CF=84,_?=\r\n" +
				" =?utf-8?q?=CE=B7=CE=B9=CF=82_=CE=B5=CE=BE_=CF=83=CE=B1=CF=80=CE=B5=CF=81?=\r\n" +
				" =?utf-8?q?=CE=B5=CE=B8_=CE=B4=CE=B5=CF=84=CF=81=CE=B1=CF=83=CE=B8=CE=BF_?=\r\n" +
				" =?utf-8?q?=CF=83=CE=B1=CE=B5=CF=86=CE=BF=CE=BB=CE=B1,_=CE=B1=CE=B4_=CE=B4?=\r\n" +
				" =?utf-8?q?=CE=BF=CE=BB=CE=BF=CF=81_=CE=B1=CE=BB=CE=B9=CE=BA=CF=85=CE=B1?=\r\n" +
				" =?utf-8?q?=CE=BD=CE=B4=CE=BF_=CE=B7=CE=B1=CF=82._=CE=95=CF=85_=CF=80?=\r\n" +
				" =?utf-8?q?=CF=85=CF=81=CE=B8=CE=BF_=CE=B9=CF=85=CE=B4=CE=B9=CF=83=CE=BF_?=\r\n" +
				" =?utf-8?q?=CE=B5=CF=89=CF=82,_=CF=86=CE=B9=CF=83=CE=B9_=CF=83=CF=89=CE=BD?=\r\n" +
				" =?utf-8?q?=CF=83=CE=B5=CE=BA=CF=85=CE=B1=CE=B8_=CF=80=CF=81=CE=B9_=CE=B5?=\r\n" +
				" =?utf-8?q?=CF=85._=CE=91=CF=83=CF=85=CE=BC_=CF=83=CE=BF=CE=BD=CF=84=CE=B5?=\r\n" +
				" =?utf-8?q?=CE=BD=CF=84=CE=B9=CF=89=CE=BD=CE=B5=CF=82_=CE=B9=CF=85=CF=82_?=\r\n" +
				" =?utf-8?q?=CE=B5=CE=B9,_=CE=B5=CE=B9_=CE=BA=CF=85=CE=B1=CE=B5=CE=BA=CF=85?=\r\n" +
				" =?utf-8?q?=CE=B5_=CE=B9=CE=BD=CF=83=CF=89=CE=BB=CE=B5=CE=BD=CF=82_=CF=83?=\r\n" +
				" =?utf-8?q?=CE=B5=CE=BD=CF=83=CE=B9=CE=B2=CF=85=CF=82_=CE=BA=CF=85=CE=BF._?=\r\n" +
				" =?utf-8?q?=CE=95=CE=BE_=CE=BA=CF=85=CF=89=CF=84_=CE=B1=CE=BB=CE=B9=CE=B5?=\r\n" +
				" =?utf-8?q?=CE=BD=CF=85=CE=BC_=CE=B7=CE=B9=CF=82,_=CF=83=CF=85_=CF=80?=\r\n" +
				" =?utf-8?q?=CF=81=CF=89_=CF=83=CE=BF=CE=BD=CF=83=CF=85=CE=BB=CE=B1=CF=84?=\r\n" +
				" =?utf-8?q?=CF=85_=CE=BC=CE=B5=CE=B4=CE=B9=CE=BF=CF=83=CF=81=CE=B9=CE=B8?=\r\n" +
				" =?utf-8?q?=CE=B1=CE=B8=CE=B5=CE=BC._=CE=A4=CE=B9=CE=B2=CE=B9=CE=BA=CF=85?=\r\n" +
				" =?utf-8?q?=CE=B5_=CE=B9=CE=BD=CF=83=CF=84=CF=81=CF=85=CF=83=CE=B8=CE=B9?=\r\n" +
				" =?utf-8?q?=CE=BF=CF=81_=CE=BA=CF=85=CE=B9_=CE=BD=CE=BF,_=CE=B5=CF=85?=\r\n" +
				" =?utf-8?q?=CE=BC_=CE=B9=CE=B4_=CE=BA=CF=85=CE=BF=CE=B4=CF=83=CE=B9_=CF=84?=\r\n" +
				" =?utf-8?q?=CE=B1=CF=83=CE=B9=CE=BC=CE=B1=CE=B8=CE=B5=CF=82_=CE=B1=CE=B4?=\r\n" +
				" =?utf-8?q?=CF=89=CE=BB=CE=B5=CF=83=CE=B5=CE=BD=CF=82.?=")},
	}
	for _, c := range cases {
		expOffset := 32 + len(c.exp)
		if pos := bytes.LastIndex(c.exp, []byte("\r\n")); pos > -1 {
			expOffset = len(c.exp) - pos - 2
		}
		act, pos := QEncode(c.src, 32)
		if !bytes.Equal(act, c.exp) {
			t.Errorf("QEncode: got (len=%d)\n%s\nwant (len=%d)\n%s", len(act), act, len(c.exp), c.exp)
		}
		if pos != expOffset {
			t.Errorf("QEncode: got offset = %d, want %d", pos, expOffset)

		}
	}
}

var b64bmRes []byte

func benchmarkBase64Encode(srcLen int, b *testing.B) {
	src := make([]byte, srcLen)
	for i := 0; i < b.N; i++ {
		rand.Read(src)
		b64bmRes = Base64Encode(src)
	}
}

func benchmarkBase64Encode_stdlib(srcLen int, b *testing.B) {
	src := make([]byte, srcLen)
	for i := 0; i < b.N; i++ {
		rand.Read(src)
		el := base64.StdEncoding.EncodedLen(srcLen)
		b64 := make([]byte, el)
		base64.StdEncoding.Encode(b64, src)
		exp := make([]byte, 0, el+el/76*2)
		for len(b64) > 76 {
			exp = append(exp, b64[:76]...)
			exp = append(exp, '\r', '\n')
			b64 = b64[76:]
		}
		b64bmRes = append(exp, b64...)
	}
}

func Benchmark_Base64Encode_1k(b *testing.B) {
	benchmarkBase64Encode(1024, b)
}

func Benchmark_Base64Encode_stdlib_1k(b *testing.B) {
	benchmarkBase64Encode_stdlib(1024, b)
}

func Benchmark_Base64Encode_4k(b *testing.B) {
	benchmarkBase64Encode(4096, b)
}

func Benchmark_Base64Encode_stdlib_4k(b *testing.B) {
	benchmarkBase64Encode_stdlib(4096, b)
}

func Benchmark_Base64Encode_10k(b *testing.B) {
	benchmarkBase64Encode(10240, b)
}

func Benchmark_Base64Encode_stdlib_10k(b *testing.B) {
	benchmarkBase64Encode_stdlib(10240, b)
}

func Benchmark_Base64Encode_40k(b *testing.B) {
	benchmarkBase64Encode(40960, b)
}

func Benchmark_Base64Encode_stdlib_40k(b *testing.B) {
	benchmarkBase64Encode_stdlib(40960, b)
}
