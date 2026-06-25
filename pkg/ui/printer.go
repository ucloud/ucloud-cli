package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/yaml.v2"
)

// Format controls the output format of a Printer.
// The numeric values must stay in sync with pkg/cli.OutputFormat (Table=0, JSON=1, YAML=2).
type Format int

const (
	Table Format = iota // 0
	JSON                // 1
	YAML                // 2
)

// Printer renders structured data to Out in the requested Format.
type Printer struct {
	Out    io.Writer
	Format Format
}

// PrintList renders dataSet (a slice or array of structs) to p.Out.
// For Table: derives column names from the first element's field names.
// For JSON/YAML: serializes the whole dataSet.
// Non-slice/array input with Table format renders nothing (mirrors base.PrintTableS behaviour).
func (p Printer) PrintList(dataSet interface{}) {
	switch p.Format {
	case JSON:
		b, err := json.MarshalIndent(dataSet, "", "  ")
		if err != nil {
			return
		}
		fmt.Fprintln(p.Out, string(b))
	case YAML:
		b, err := yaml.Marshal(dataSet)
		if err != nil {
			return
		}
		_, _ = p.Out.Write(b)
	default: // Table
		val := reflect.ValueOf(dataSet)
		fieldNameList := make([]string, 0)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			if val.Len() > 0 {
				elemType := val.Index(0).Type()
				for i := 0; i < elemType.NumField(); i++ {
					fieldNameList = append(fieldNameList, elemType.Field(i).Name)
				}
			}
			displaySlice(p.Out, val, fieldNameList)
		}
	}
}

// PrintJSON renders dataSet as indented JSON to out.
func PrintJSON(dataSet interface{}, out io.Writer) error {
	b, err := json.MarshalIndent(dataSet, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(b))
	if err != nil {
		return err
	}
	return nil
}

// gap is the number of spaces between table columns.
const gap = 2

// calcCutWidth counts extra display-width consumed by CJK/non-Latin punctuation
// (each such rune occupies 2 terminal cells but len() counts 3 bytes, so the
// difference — 1 extra cell per rune — must be subtracted when padding).
func calcCutWidth(text string) int {
	set := []*unicode.RangeTable{unicode.Han, unicode.Punct}
	width := 0
	for _, r := range text {
		if unicode.IsOneOf(set, r) && r > unicode.MaxLatin1 {
			width++
		}
	}
	return width
}

// calcWidth returns the terminal display width of text,
// counting CJK/non-Latin punctuation as 2 cells and ASCII as 1.
func calcWidth(text string) int {
	set := []*unicode.RangeTable{unicode.Han, unicode.Punct}
	width := 0
	for _, r := range text {
		if unicode.IsOneOf(set, r) && r > unicode.MaxLatin1 {
			width += 2
		} else {
			width++
		}
	}
	return width
}

func displaySlice(out io.Writer, listVal reflect.Value, fieldList []string) {
	showFieldMap := make(map[string]int)
	for _, field := range fieldList {
		showFieldMap[field] = len([]rune(field))
	}
	rowList := make([]map[string]interface{}, 0)
	for i := 0; i < listVal.Len(); i++ {
		elemVal := listVal.Index(i)
		elemType := elemVal.Type()
		var rows []map[string]interface{}
		for j := 0; j < elemVal.NumField(); j++ {
			field := elemVal.Field(j)
			fieldName := elemType.Field(j).Name
			if _, ok := showFieldMap[fieldName]; ok {
				if field.Kind() == reflect.Ptr {
					field = field.Elem()
				}
				text := fmt.Sprintf("%v", field.Interface())
				cells := strings.Split(text, "\n")
				for i, cell := range cells {
					width := calcWidth(cell)
					if showFieldMap[fieldName] < width {
						showFieldMap[fieldName] = width
					}
					if len(rows) == i {
						rows = append(rows, make(map[string]interface{}))
					}
					rows[i][fieldName] = cell
				}
			}
		}
		rowList = append(rowList, rows...)
	}
	printTable(out, rowList, fieldList, showFieldMap)
}

func printTable(out io.Writer, rowList []map[string]interface{}, fieldList []string, fieldWidthMap map[string]int) {
	for _, field := range fieldList {
		tmpl := "%-" + strconv.Itoa(fieldWidthMap[field]+gap) + "s"
		fmt.Fprintf(out, tmpl, field)
	}
	if len(fieldList) != 0 {
		fmt.Fprintf(out, "\n")
	}
	for _, row := range rowList {
		for _, field := range fieldList {
			cutWidth := calcCutWidth(fmt.Sprintf("%v", row[field]))
			tmpl := "%-" + strconv.Itoa(fieldWidthMap[field]-cutWidth+gap) + "v"
			if row[field] != nil {
				fmt.Fprintf(out, tmpl, row[field])
			} else {
				fmt.Fprintf(out, tmpl, "")
			}
		}
		fmt.Fprintf(out, "\n")
	}
}
