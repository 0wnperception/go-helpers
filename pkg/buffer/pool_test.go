package buffer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendByte(t *testing.T) {
	var b Buffer
	var want []byte

	for i := 0; i < 1000; i++ {
		b.AppendByte(1)
		b.AppendByte(2)
		want = append(want, 1, 2)
	}

	got := b.BuildBytes()
	if !bytes.Equal(got, want) {
		t.Errorf("BuildBytes() = %v; want %v", got, want)
	}
}

func TestEnsureSpace(t *testing.T) {
	var b Buffer

	b.EnsureSpace(1023)

	if cap(b.Buf) != 1024 {
		t.Error("Invalid length")
	}

	b = Buffer{}

	b.EnsureSpace(1025)
	if cap(b.Buf) != 2048 {
		t.Error("Invalid length")
	}

	b = Buffer{}

	b.EnsureSpace(1025000000)
	if cap(b.Buf) != 32768 {
		t.Error("Invalid length")
	}
}

func TestAppendBytes(t *testing.T) {
	var b Buffer
	var want []byte

	for i := 0; i < 1000; i++ {
		b.AppendBytes([]byte{1, 2})
		want = append(want, 1, 2)
	}

	got := b.BuildBytes()
	if !bytes.Equal(got, want) {
		t.Errorf("BuildBytes() = %v; want %v", got, want)
	}
}

func TestAppendString(t *testing.T) {
	var b Buffer
	var want []byte

	s := "test"
	for i := 0; i < 1000; i++ {
		b.AppendBytes([]byte(s))
		want = append(want, s...)
	}

	got := b.BuildBytes()
	if !bytes.Equal(got, want) {
		t.Errorf("BuildBytes() = %v; want %v", got, want)
	}
}

func TestDumpTo(t *testing.T) {
	var b Buffer
	var want []byte

	s := "test"
	for i := 0; i < 1000; i++ {
		b.AppendBytes([]byte(s))
		want = append(want, s...)
	}

	out := &bytes.Buffer{}
	n, err := b.WriteTo(out)
	if err != nil {
		t.Errorf("DumpTo() error: %v", err)
	}

	got := out.Bytes()
	if !bytes.Equal(got, want) {
		t.Errorf("DumpTo(): got %v; want %v", got, want)
	}

	if n != int64(len(want)) {
		t.Errorf("DumpTo() = %v; want %v", n, len(want))
	}
}

func TestReadCloser(t *testing.T) {
	var b Buffer
	var want []byte

	s := "test"
	for i := 0; i < 1000; i++ {
		b.AppendBytes([]byte(s))
		want = append(want, s...)
	}

	out := &bytes.Buffer{}
	rc := b.ReadCloser()
	n, err := out.ReadFrom(rc)
	if err != nil {
		t.Errorf("ReadCloser() error: %v", err)
	}
	_ = rc.Close() // Will always return nil

	got := out.Bytes()
	if !bytes.Equal(got, want) {
		t.Errorf("DumpTo(): got %v; want %v", got, want)
	}

	if n != int64(len(want)) {
		t.Errorf("DumpTo() = %v; want %v", n, len(want))
	}
}

func TestAppendXmlEncode(t *testing.T) {
	b := Buffer{}
	b.AppendXMLEncode(`q< " > ' &`)

	s := string(b.BuildBytes())
	if s != `q&lt; &quot; &gt; &apos; &amp;` {
		t.Errorf("Invalid xml encode")
	}
}

func TestAppendXmlElement(t *testing.T) {
	b := Buffer{}

	b.AppendXMLElement("abc", "v<a>lu&e")

	s := string(b.BuildBytes())
	if s != "<abc>v&lt;a&gt;lu&amp;e</abc>" {
		t.Errorf("Invalid appended xml element, %v", s)
	}
}

func TestAppendStrings(t *testing.T) {
	b := Buffer{}

	b.AppendStrings("a", "bc", "def")
	s := string(b.BuildBytes())
	if s != "abcdef" {
		t.Errorf("Invalid appended string array, %v", s)
	}
}

const readStr = `iwoefu[qwpoiefm[pqeowifm qpeifom q[peifmcq[pweimfc]qp[weimfv]q[pweimfv]qw[pemvfi]q[wpoevfmq][wepofv
qweounq[pwoevrmq[wpeoimrvq[pweoimrvq[pweriomvq[pwrimevq[peirmvq[pweimrvq[pweirvmq[pweirmvp[qweimrvq[pweirmvq[pweirmv
qwnveurw[opuev[pqowrmevcqp[owimvrMPWEOIVMF]W[EPIMRVP]QEIMVFR][WEPIRMV]W[EPIMRV][WEPIMR ][WEPIRM V]W[REPVM[EPRVMOW][ERFV
QWOPEUNQ[POWEUV[PWOEIFM PWQEOIM VQPOEIMF QPEOFI MQW][PFEIMV ]PQW[MIF][WEPO MPEQWFI M]Q[WE MPQW[EI MQPOWEIMVPQEI V[PQIWEцзум
цмузкшм хукзшмхцшумхцзукщьмцъхХЦЗУЬМКХЪЗЦУЗЩЬЕМХЗЦУЗЩКБМХЦУКЗЩБМЦУХКЗЩМБЦУ
ХЪКЩМБЦХУКЪЕЩМБЦХУЪЗКЩБМХЪКЦУЩБЕМХЗКУЦЩЕБМКЗХУЦЕЩИЬЗЦКШЕИЬЦУЪКХЗШЕ ХУКЕЩИБХЦЪУЩЬКХЦЪКУЗЩЕБМЦХЗКeo'rimgw[e]prmvgw[eeprovm
wer;iomvw'eprimfvw]e[rpfmoviw';erimvw]erp,fvw[]pero,v]w[epro,vr[ewpo,v]'ewpro,v][epwrp,ov][erw,ovg[wer,ov[wero,v[]ewrogv,
we[prv-459v[weovm[]wve-rm9tv]we[,o[eprg,vew],rovge[wr0]v
]ervp,g[]repwv,gre
][p,g
tre][bp.ger]
p][eprvg]e[rpgvrt[]bt[rb,otr'[b,ort'[oh,bter],boh
rt][,bhr
]e'`

func TestRead(t *testing.T) {
	b := Buffer{}

	i, err := b.read(strings.NewReader(readStr), 256)

	require.NoError(t, err)
	require.Equal(t, int64(len(readStr)), i)
	require.Equal(t, readStr, string(b.BuildBytes(nil)))
}
