// Package wordbook reads local XLSX word lists without Excel or network access.
package wordbook

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Word struct {
	English            string   `json:"english"`
	Phonetic           string   `json:"phonetic"`
	Meaning            string   `json:"meaning"`
	ChineseDistractors []string `json:"chineseDistractors"`
	EnglishDistractors []string `json:"englishDistractors"`
}

type richText struct {
	Text string `xml:"t"`
	Runs []struct {
		Text string `xml:"t"`
	} `xml:"r"`
}

func (s richText) value() string {
	var b strings.Builder
	b.WriteString(s.Text)
	for _, r := range s.Runs {
		b.WriteString(r.Text)
	}
	return strings.TrimSpace(b.String())
}

type cell struct {
	Ref    string   `xml:"r,attr"`
	Kind   string   `xml:"t,attr"`
	Value  string   `xml:"v"`
	Inline richText `xml:"is"`
}

// WriteTemplate includes optional, author-supplied Chinese quiz distractors.
func WriteTemplate(filename string) error {
	headers := []string{"英文单词", "音标", "英文单词释义", "常用短语", "中文干扰项1", "中文干扰项2", "中文干扰项3"}
	var sheet strings.Builder
	sheet.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><cols><col min="1" max="2" width="20" customWidth="1"/><col min="3" max="3" width="60" customWidth="1"/><col min="4" max="4" width="50" customWidth="1"/><col min="5" max="7" width="28" customWidth="1"/></cols><sheetData><row r="1">`)
	for i, header := range headers {
		fmt.Fprintf(&sheet, `<c r="%c1" t="inlineStr"><is><t>`, 'A'+i)
		if err := xml.EscapeText(&sheet, []byte(header)); err != nil {
			return err
		}
		sheet.WriteString(`</t></is></c>`)
	}
	sheet.WriteString(`</row></sheetData></worksheet>`)
	files := []struct{ name, body string }{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="单词表" sheetId="1" r:id="rId1"/></sheets></workbook>`},
		{"xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`},
		{"xl/worksheets/sheet1.xml", sheet.String()},
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(filename), ".fuckenglish-template-*.xlsx")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	z := zip.NewWriter(tmp)
	for _, file := range files {
		w, err := z.Create(file.name)
		if err != nil {
			z.Close()
			tmp.Close()
			return err
		}
		if _, err := io.WriteString(w, file.body); err != nil {
			z.Close()
			tmp.Close()
			return err
		}
	}
	if err := z.Close(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filename)
}

// Read uses the first worksheet, resolving its relationship rather than assuming sheet1.xml.
func Read(filename string) ([]Word, error) {
	z, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开 Excel：%w", err)
	}
	defer z.Close()
	files := map[string]*zip.File{}
	var size uint64
	for _, f := range z.File {
		size += f.UncompressedSize64
		if size > 64<<20 {
			return nil, errors.New("Excel 解压后超过 64 MB，请拆分词表")
		}
		files[f.Name] = f
	}
	read := func(name string, dst any) error {
		f := files[name]
		if f == nil {
			return fmt.Errorf("Excel 缺少 %s", name)
		}
		r, e := f.Open()
		if e != nil {
			return e
		}
		defer r.Close()
		return xml.NewDecoder(io.LimitReader(r, 32<<20)).Decode(dst)
	}
	var book struct {
		Sheets []struct {
			ID string `xml:"id,attr"`
		} `xml:"sheets>sheet"`
	}
	if err := read("xl/workbook.xml", &book); err != nil {
		return nil, err
	}
	if len(book.Sheets) == 0 {
		return nil, errors.New("Excel 没有工作表")
	}
	var rels struct {
		Items []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := read("xl/_rels/workbook.xml.rels", &rels); err != nil {
		return nil, err
	}
	sheetPath := ""
	for _, r := range rels.Items {
		if r.ID == book.Sheets[0].ID {
			if strings.HasPrefix(r.Target, "/") {
				sheetPath = strings.TrimPrefix(path.Clean(r.Target), "/")
			} else {
				sheetPath = path.Join("xl", r.Target)
			}
		}
	}
	if sheetPath == "" {
		return nil, errors.New("无法找到第一个工作表")
	}
	var shared struct {
		Items []richText `xml:"si"`
	}
	if files["xl/sharedStrings.xml"] != nil {
		if err := read("xl/sharedStrings.xml", &shared); err != nil {
			return nil, err
		}
	}
	var sheet struct {
		Rows []struct {
			Cells []cell `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := read(sheetPath, &sheet); err != nil {
		return nil, fmt.Errorf("读取工作表失败：%w", err)
	}
	text := func(c cell) string {
		if c.Kind == "inlineStr" {
			return c.Inline.value()
		}
		if c.Kind == "s" {
			var i int
			if _, e := fmt.Sscan(c.Value, &i); e == nil && i >= 0 && i < len(shared.Items) {
				return shared.Items[i].value()
			}
			return ""
		}
		return strings.TrimSpace(c.Value)
	}
	columns := map[string]int{}
	header := false
	words := []Word{}
	for _, row := range sheet.Rows {
		values := map[int]string{}
		for i, c := range row.Cells {
			col := 0
			for _, r := range c.Ref {
				if r < 'A' || r > 'Z' {
					break
				}
				col = col*26 + int(r-'A'+1)
			}
			if col == 0 {
				col = i + 1
			}
			values[col-1] = text(c)
		}
		if !header {
			for col, v := range values {
				v = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(v, "\ufeff")))
				switch v {
				case "英文单词", "单词", "英文", "word", "english":
					columns["english"] = col
				case "音标", "phonetic", "ipa", "pronunciation":
					columns["phonetic"] = col
				case "英文单词释义", "中文释义", "中文翻译", "释义", "翻译", "meaning", "translation":
					columns["meaning"] = col
				}
				for _, direction := range []string{"中文", "英文"} {
					for i := 1; i <= 3; i++ {
						key := fmt.Sprintf("%s干扰项%d", direction, i)
						if v == key {
							columns[key] = col
						}
					}
				}
			}
			_, english := columns["english"]
			_, phonetic := columns["phonetic"]
			_, meaning := columns["meaning"]
			if english && phonetic && meaning {
				header = true
			}
			continue
		}
		w := Word{English: values[columns["english"]], Phonetic: values[columns["phonetic"]], Meaning: values[columns["meaning"]], ChineseDistractors: make([]string, 3), EnglishDistractors: make([]string, 3)}
		for i := 0; i < 3; i++ {
			if col, ok := columns[fmt.Sprintf("中文干扰项%d", i+1)]; ok {
				w.ChineseDistractors[i] = values[col]
			}
			if col, ok := columns[fmt.Sprintf("英文干扰项%d", i+1)]; ok {
				w.EnglishDistractors[i] = values[col]
			}
		}
		if w.English != "" {
			if w.Meaning == "" {
				return nil, fmt.Errorf("单词 %q 缺少中文释义", w.English)
			}
			words = append(words, w)
		}
		if len(words) > 50000 {
			return nil, errors.New("单份词表最多支持 50000 个词，请拆分文件")
		}
	}
	if !header {
		return nil, errors.New("首个工作表需要表头：英文单词、音标、英文单词释义")
	}
	if len(words) == 0 {
		return nil, errors.New("词表没有可读取的单词")
	}
	return words, nil
}
