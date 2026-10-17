package editor

import (
	"encoding/xml"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

func TestXMLCommentFixerReplacesDashesInComments(t *testing.T) {
	input := "<!--- a -- b --->x<![CDATA[<!-- -- -->]]>y<!---->z"
	want := "<!--  a    b  -->x<![CDATA[<!-- -- -->]]>y<!---->z"

	got, err := io.ReadAll(newXMLCommentFixer(strings.NewReader(input)))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

// fixCommentsSlowly is the same fix done in the simplest way, to check reads split at any place
func fixCommentsSlowly(s string) string {
	out := []byte(s)
	state := xmlFixerText
	for i := 0; i < len(out); i++ {
		rest := s[i:]
		switch {
		case state == xmlFixerText && strings.HasPrefix(rest, "<!--"):
			state, i = xmlFixerComment, i+3
		case state == xmlFixerText && strings.HasPrefix(rest, "<![CDATA["):
			state, i = xmlFixerCData, i+8
		case state == xmlFixerComment && strings.HasPrefix(rest, "-->"),
			state == xmlFixerCData && strings.HasPrefix(rest, "]]>"):
			state, i = xmlFixerText, i+2
		case state == xmlFixerComment && out[i] == '-':
			out[i] = ' '
		}
	}
	return string(out)
}

func TestXMLCommentFixerSplitReads(t *testing.T) {
	var sb strings.Builder
	for i := 0; sb.Len() < 100*1024; i++ {
		// Different lengths move the sequences across the buffer boundary
		sb.WriteString(strings.Repeat("x", i%11))
		sb.WriteString("<!--- a -- b ---><A>-- ]]></A><![CDATA[ <!-- ]] ]]]><!---->")
	}
	input := sb.String()
	want := fixCommentsSlowly(input)

	readers := map[string]func() io.Reader{
		"whole":    func() io.Reader { return strings.NewReader(input) },
		"one byte": func() io.Reader { return iotest.OneByteReader(strings.NewReader(input)) },
		"half":     func() io.Reader { return iotest.HalfReader(strings.NewReader(input)) },
		"data+EOF": func() io.Reader { return iotest.DataErrReader(strings.NewReader(input)) },
	}
	for name, reader := range readers {
		got, err := io.ReadAll(newXMLCommentFixer(reader()))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s: output differs", name)
		}
		got, err = io.ReadAll(iotest.OneByteReader(newXMLCommentFixer(reader())))
		if err != nil || string(got) != want {
			t.Errorf("%s: output differs when read by one byte (%v)", name, err)
		}
	}
}

func TestXMLCommentFixerReturnsReadErrors(t *testing.T) {
	_, err := io.ReadAll(newXMLCommentFixer(iotest.TimeoutReader(strings.NewReader("<A><!-- -- --></A>"))))
	if !errors.Is(err, iotest.ErrTimeout) {
		t.Errorf("read error must be returned, got %v", err)
	}
}

func TestXMLDecoderAcceptsDashesInComments(t *testing.T) {
	input := `<?xml version="1.0" encoding="ISO-8859-1"?>
<!-- ---------------------------------------- -->
<!--- Dale's Combat Mod --->
<Root>
	<!-- DCM -- start -->
	<A>x--y</A>
	<B><![CDATA[<!-- not -- a comment -->]]></B>
	<C>caf` + "\xe9" + `</C>
	<!---->
</Root>`

	var root struct {
		A string
		B string
		C string
	}
	if err := newXMLDecoder(strings.NewReader(input)).Decode(&root); err != nil {
		t.Fatal(err)
	}
	if root.A != "x--y" || root.B != "<!-- not -- a comment -->" || root.C != "café" {
		t.Errorf("text changed: %+v", root)
	}
}

func TestXMLDecoderKeepsLineNumbers(t *testing.T) {
	input := "<Root>\n<!-- a -- b -->\n<!- broken\n</Root>"

	err := newXMLDecoder(strings.NewReader(input)).Decode(&struct{}{})
	var syntaxErr *xml.SyntaxError
	if !errors.As(err, &syntaxErr) || syntaxErr.Line != 3 {
		t.Errorf("syntax error on line 3 expected, got %v", err)
	}
}

func TestLoadXMLFilesWithDashesInComments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CIV4PromotionInfos.xml")
	writeTestFile(t, path, `<?xml version="1.0"?>
<!-- Sid Meier's Civilization 4 -->
<!--- Chronicles of Mankind --->
<Civ4TechInfos>
	<TechInfos>
		<!-- ---------- Ancient ---------- -->
		<TechInfo><Type>TECH_MINING</Type><Description>TXT_KEY_TECH_MINING</Description></TechInfo>
	</TechInfos>
</Civ4TechInfos>`)

	data := LoadXMLFiles([]string{path}, nil)
	if data.Table(InfoTechs).Get("TECH_MINING") == nil {
		t.Error("file with dashes in comments must be loaded")
	}
}
