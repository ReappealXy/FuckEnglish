package wordbook

import (
	"archive/zip"
	"os"
	"path/filepath"
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
