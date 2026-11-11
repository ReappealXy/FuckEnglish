package wordbook

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRealWordLists(t *testing.T) {
	for _, name := range []string{"list1.xlsx", "list2.xlsx"} {
		t.Run(name, func(t *testing.T) {
			words, err := Read(filepath.Join("..", "..", "words", name))
			if err != nil {
				t.Fatal(err)
			}
			if len(words) != 50 {
				t.Fatalf("got %d words", len(words))
			}
			for _, w := range words {
				if w.English == "" || w.Phonetic == "" || w.Meaning == "" {
					t.Fatalf("incomplete word: %+v", w)
				}
			}
		})
	}
}

func TestWriteTemplate(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "template.xlsx")
	if err := WriteTemplate(filename); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var sheet string
	for _, f := range z.File {
		if f.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		sheet = string(b)
	}
	var parsed struct {
		Rows []struct {
			Cells []cell `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal([]byte(sheet), &parsed); err != nil {
		t.Fatal(err)
	}
	headers := []string{"英文单词", "音标", "英文单词释义", "常用短语", "中文干扰项1", "中文干扰项2", "中文干扰项3"}
	if len(parsed.Rows) != 1 || len(parsed.Rows[0].Cells) != len(headers) {
		t.Fatalf("expected an empty template with exactly seven headers: %s", sheet)
	}
	for i, header := range headers {
		c := parsed.Rows[0].Cells[i]
		wantRef := string(rune('A'+i)) + "1"
		if c.Ref != wantRef || c.Kind != "inlineStr" || c.Inline.value() != header {
			t.Errorf("header %s: got %+v, want %q", wantRef, c, header)
		}
	}
}

func fixture(t *testing.T, shared, sheet string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "list.xlsx")
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	files := map[string]string{"xl/workbook.xml": `<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet r:id="rId7"/></sheets></workbook>`, "xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId7" Target="/xl/worksheets/custom.xml"/></Relationships>`, "xl/worksheets/custom.xml": sheet}
	if shared != "" {
		files["xl/sharedStrings.xml"] = shared
	}
	for name, s := range files {
		w, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write([]byte(s)); e != nil {
			t.Fatal(e)
		}
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	f.Close()
	return filename
}
func TestSharedStringsAndSparseCells(t *testing.T) {
	f := fixture(t, `<sst><si><t>英文单词</t></si><si><t>音标</t></si><si><t>英文单词释义</t></si><si><r><t>pio</t></r><r><t>neer</t></r></si><si><t>先驱者</t></si></sst>`, `<worksheet><sheetData><row><c r="B1" t="s"><v>0</v></c><c r="D1" t="s"><v>1</v></c><c r="F1" t="s"><v>2</v></c></row><row><c r="B2" t="s"><v>3</v></c><c r="F2" t="s"><v>4</v></c></row></sheetData></worksheet>`)
	w, e := Read(f)
	if e != nil {
		t.Fatal(e)
	}
	if len(w) != 1 || w[0].English != "pioneer" || w[0].Meaning != "先驱者" || w[0].Phonetic != "" {
		t.Fatalf("bad word: %+v", w)
	}
}
func TestRejectMissingHeadersAndCorruption(t *testing.T) {
	f := fixture(t, "", `<worksheet><sheetData><row><c r="A1" t="inlineStr"><is><t>hello</t></is></c></row></sheetData></worksheet>`)
	if _, e := Read(f); e == nil {
		t.Fatal("missing headers accepted")
	}
	bad := filepath.Join(t.TempDir(), "bad.xlsx")
	os.WriteFile(bad, []byte("not Excel"), 0600)
	if _, e := Read(bad); e == nil {
		t.Fatal("corrupt file accepted")
	}
}

func TestOptionalDistractorsWithReorderedSparseColumns(t *testing.T) {
	headers := []string{"英文干扰项3", "中文干扰项2", "英文单词", "中文释义", "中文干扰项1", "音标", "英文干扰项1", "中文干扰项3", "英文干扰项2"}
	values := []string{"runway", "飞行员", "aircraft", "航空器", "机场", "", "airport", "跑道", "pilot"}
	var sheet strings.Builder
	sheet.WriteString(`<worksheet><sheetData>`)
	for row, cells := range [][]string{headers, values} {
		fmt.Fprintf(&sheet, `<row r="%d">`, row+1)
		for col, value := range cells {
			fmt.Fprintf(&sheet, `<c r="%c%d" t="inlineStr"><is><t>%s</t></is></c>`, 'B'+col*2, row+1, value)
		}
		sheet.WriteString(`</row>`)
	}
	sheet.WriteString(`</sheetData></worksheet>`)
	words, err := Read(fixture(t, "", sheet.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 1 || words[0].English != "aircraft" || !reflect.DeepEqual(words[0].ChineseDistractors, []string{"机场", "飞行员", "跑道"}) || !reflect.DeepEqual(words[0].EnglishDistractors, []string{"airport", "pilot", "runway"}) {
		t.Fatalf("parsed: %+v", words)
	}
	for _, filename := range []string{"list1.xlsx", "list2.xlsx"} {
		words, err := Read(filepath.Join("..", "..", "words", filename))
		if err != nil || len(words) != 50 || len(words[0].ChineseDistractors) != 3 || words[0].ChineseDistractors[0] != "" {
			t.Fatalf("old workbook no longer browsable: %v", err)
		}
	}
}
