package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/xuri/excelize/v2"
)

type TestResultEntry struct {
	Time    time.Time
	Action  string
	Package string
	Output  string
	Elapsed float32
}

// Outputs
// - CSV file
// - SQL Lite
// - Excel Tabs
// Report
// - Summary
// - Charts

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			runTestCmd(),
			runOutputsCmd(),
			runSummarize(),
		},
		Name:                 "Arkloud Test CLI",
		EnableBashCompletion: true,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("ERROR : %v\n", err)
	}

}

func GetGitHash() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return ""
	}

	result := string(output)
	return strings.TrimSpace(result)
}

func SaveResults(results []*TestResultEntry, runId string, outputs []string) {
	if len(outputs) == 0 {
		outputs = []string{"xlsx"}
	}

	for _, o := range outputs {
		switch o {
		case "sqllite":
			SaveAsSqlLite(results, runId, "")
		case "csv":
			SaveAsCSV(results, runId, "")
		case "xlsx":
			SaveAsExcel(results, runId, "")
		}
	}
}

func ParseGoTestJson(r io.Reader) []*TestResultEntry {
	var results []*TestResultEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		txt := scanner.Text()
		line := &TestResultEntry{}
		err := json.Unmarshal([]byte(txt), line)
		if err != nil {
			log.Fatalf("Error: Bad line: %v\n", err)
		}
		if line.Action == "pass" || line.Action == "fail" {
			results = append(results, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return results
}

func SaveAsSqlLite(results []*TestResultEntry, runId string, filename string) {

}

func SaveAsCSV(results []*TestResultEntry, runId string, csvFile string) {
	var err error
	if csvFile == "" {
		csvFile = "test-results.csv"
	}

	exists, err := Exists(csvFile)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	var file *os.File
	if exists {
		file, err = os.OpenFile(csvFile, os.O_APPEND, 0600)
	} else {
		file, err = os.OpenFile(csvFile, os.O_CREATE, 0600)
	}
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	if !exists {
		// Write header
		rec := []string{
			"RunId",
			"Time",
			"Action",

			"Package",
			"Elapsed",
		}
		err = w.Write(rec)
		if err != nil {
			fmt.Printf("Error Bad Line %v\n", err)
		}
	}

	for _, line := range results {
		rec := []string{
			runId,
			line.Time.String(),
			line.Action,
			line.Package,
			fmt.Sprintf("%v", line.Elapsed),
		}
		err = w.Write(rec)
		if err != nil {
			fmt.Printf("Error Bad Line %v\n", err)
		}
	}
	w.Flush()
}

func GenerateReport() {

}

func SortKeys[T any](m map[string]T) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	return keys
}

func GetCommonPrefix[T any](m map[string]T) string {
	p := ""
	for i := 1; i < 500; i++ {
		first := ""
		for k := range m {
			if first == "" {
				first = k[:i]
				continue
			}
			if first != k[:i] {
				return p
			}
		}
		p = first
	}
	return p
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func runTestCmd() *cli.Command {
	return &cli.Command{
		Name:  "gotest",
		Usage: "gotest <dir> -o sqllite -o csv -o stdout",
		Action: func(c *cli.Context) error {
			cmd := exec.Command("go", "test", "./...", "-json")
			// cmd := exec.Command("go", "test -coverprofile=cover.out -json ./...")
			b := new(bytes.Buffer)
			b2 := new(bytes.Buffer)
			cmd.Stdout = b
			cmd.Stderr = b2
			err := cmd.Run()
			output := b.Bytes()

			// fmt.Println(b2.String())
			// fmt.Println(b.String())

			// output, err := cmd.Output()
			// if err != nil {
			// 	fmt.Println("Error executing command:", err)
			// 	return err
			// }

			// Save the results
			rawFile := c.String("raw")
			if rawFile != "" {

				err = os.WriteFile(rawFile, output, 0600)
				if err != nil {
					fmt.Println("Error Saving Rav file:", err)
					return err
				}
			}

			// Parse the results
			r := bytes.NewBuffer(output)
			results := ParseGoTestJson(r)

			runId := GetGitHash()
			outputs := c.StringSlice("output")
			SaveResults(results, runId, outputs)

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "output",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name: "raw",
			},
		},
	}
}

func runSummarize() *cli.Command {
	return &cli.Command{
		Name:  "summarize",
		Usage: "summarize <file> -o <markdown file>",
		Action: func(c *cli.Context) error {
			fileName := c.Args().First()

			data, err := os.ReadFile(fileName)
			if err != nil {
				fmt.Println("Error parsing file:", err)
				return err
			}
			lines := strings.Split(string(data), "\n")

			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, fileName, nil, parser.AllErrors)
			if err != nil {
				fmt.Println("Error parsing file:", err)
				return err
			}

			var funcs []fnRec

			for _, decl := range node.Decls {
				if fn, isFn := decl.(*ast.FuncDecl); isFn {
					if strings.HasPrefix(fn.Name.Name, "Test") {
						start := fset.Position(fn.Pos()).Line
						end := fset.Position(fn.End()).Line
						funcs = append(funcs, fnRec{
							name:  fn.Name.Name,
							start: start,
							end:   end,
						})
					}
				}
			}

			for _, fr := range funcs {
				fr.getComments(lines)
				fr.PrintMD()
			}

			return nil
		},
		Args: true,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Required: false,
			},
		},
	}
}

type fnRec struct {
	name      string
	start     int
	end       int
	fnComment string
	comments  []string
}

func (rec *fnRec) getComments(lines []string) {
	// Get the function level comments
	var fnCommentLines []string

	for i := rec.start - 2; i >= 0; i-- {
		c := rec.GetComment(lines[i])
		if c == "" {
			break
		}
		fnCommentLines = append(fnCommentLines, c)
	}
	slices.Reverse(fnCommentLines)
	rec.fnComment = strings.Join(fnCommentLines, " ")

	var current []string
	for i := rec.start + 1; i < rec.end; i++ {
		comment := rec.GetComment(lines[i])
		if comment == "" {
			if len(current) > 0 {
				c := strings.Join(current, " ")
				rec.comments = append(rec.comments, c)
				current = []string{}
			}
			continue
		}
		current = append(current, comment)
	}
}

func (rec *fnRec) PrintMD() {
	name := rec.camelCaseToSpaces(rec.name[4:])
	fmt.Printf("## %v\n", name)
	fmt.Printf("\n")
	fmt.Printf("%v\n", rec.fnComment)
	fmt.Printf("\n")
	for i, comment := range rec.comments {
		fmt.Printf("%v. %v\n", i+1, comment)
	}
}

func (rec *fnRec) camelCaseToSpaces(str string) string {
	// Use a regex to identify where to add spaces
	regex := regexp.MustCompile("([a-z])([A-Z])")
	withSpaces := regex.ReplaceAllString(str, "${1} ${2}")
	return withSpaces
}

func (rec *fnRec) GetComment(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") {
		return strings.TrimSpace(trimmed[2:])
	}
	return ""
}

func runOutputsCmd() *cli.Command {
	return &cli.Command{
		Name:  "output",
		Usage: "output <testresults> -o sqllite -o csv -o stdout",
		Action: func(c *cli.Context) error {
			outputs := c.StringSlice("output")

			outputFile := c.Args().First()
			output, err := os.ReadFile(outputFile)
			if err != nil {
				fmt.Println("Error executing command:", err)
				return err
			}

			// Parse the results
			r := bytes.NewBuffer(output)
			results := ParseGoTestJson(r)

			runId := GetGitHash()
			SaveResults(results, runId, outputs)

			return nil
		},
		Args: true,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Required: true,
			},
		},
	}
}

func SaveAsExcel(results []*TestResultEntry, runId string, filename string) {

	var err error
	if filename == "" {
		filename = "test-results.xlsx"
	}

	exists, err := Exists(filename)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	var file *excelize.File
	if !exists {
		file = excelize.NewFile()
	} else {
		file, err = excelize.OpenFile(filename)
	}
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	sheetName := runId[:30]

	// _, ok := file.Sheet.Load(sheetName)
	// if ok {
	err = file.DeleteSheet(sheetName)
	// }

	die2(file.NewSheet(sheetName))

	// Generate Raw data
	die(file.SetCellStr(sheetName, "A1", "RunId"))
	die(file.SetCellStr(sheetName, "B1", "Time"))
	die(file.SetCellStr(sheetName, "C1", "Action"))
	die(file.SetCellStr(sheetName, "D1", "Package"))
	die(file.SetCellStr(sheetName, "E1", "Elapsed"))

	passFail := make(map[string]int)
	packageResults := make(map[string][]int)

	for i, line := range results {
		row := i + 2
		die(file.SetCellValue(sheetName, fmt.Sprintf("A%v", row), runId))
		die(file.SetCellValue(sheetName, fmt.Sprintf("B%v", row), line.Time))
		die(file.SetCellValue(sheetName, fmt.Sprintf("C%v", row), line.Action))
		die(file.SetCellValue(sheetName, fmt.Sprintf("D%v", row), line.Package))
		die(file.SetCellValue(sheetName, fmt.Sprintf("E%v", row), line.Elapsed))

		passFail[line.Action] = passFail[line.Action] + 1
		pkg := packageResults[line.Package]
		if pkg == nil {
			pkg = make([]int, 2)
			packageResults[line.Package] = pkg
		}
		if line.Action == "pass" {
			pkg[0] = pkg[0] + 1
		} else if line.Action == "fail" {
			pkg[1] = pkg[1] + 1
		}
	}

	// Generate Summary Data
	die(file.SetCellStr(sheetName, "G1", "All"))
	die(file.SetCellStr(sheetName, "H1", "Pass"))
	die(file.SetCellStr(sheetName, "I1", "Fail"))
	die(file.SetCellValue(sheetName, "H2", passFail["pass"]))
	die(file.SetCellValue(sheetName, "I2", passFail["fail"]))
	formula := fmt.Sprintf("=H%v/(I%v+H%v)", 2, 2, 2)
	die(file.SetCellFormula(sheetName, fmt.Sprintf("J%v", 2), formula))

	// Generate the Pass / Fail Pie chart
	die(file.SetCellStr(sheetName, "G4", "Package"))
	die(file.SetCellStr(sheetName, "H4", "Pass"))
	die(file.SetCellStr(sheetName, "I4", "Fail"))
	die(file.SetCellStr(sheetName, "J4", "% Pass"))
	row := 4

	keys := SortKeys(packageResults)
	prefix := GetCommonPrefix(packageResults)

	style, _ := file.NewStyle(&excelize.Style{
		NumFmt: 9,
	})
	for _, pkg := range keys {
		row++
		short := pkg[len(prefix):]
		result := packageResults[pkg]
		die(file.SetCellValue(sheetName, fmt.Sprintf("G%v", row), short))
		die(file.SetCellValue(sheetName, fmt.Sprintf("H%v", row), result[0]))
		die(file.SetCellValue(sheetName, fmt.Sprintf("I%v", row), result[1]))
		formula := fmt.Sprintf("=H%v/(I%v+H%v)", row, row, row)
		die(file.SetCellFormula(sheetName, fmt.Sprintf("J%v", row), formula))
	}
	die(file.SetCellStyle(sheetName, "J2", fmt.Sprintf("J%v", row), style))

	// Generate Pass Fail Pie char
	die(file.AddChart(sheetName, "L2", &excelize.Chart{
		Type: excelize.Pie,
		Title: []excelize.RichTextRun{
			{
				Text: "Pass / Fail",
			},
		},
		Series: []excelize.ChartSeries{
			{
				Name:       "",
				Values:     fmt.Sprintf("%v!$H$2:$I$2", sheetName),
				Categories: fmt.Sprintf("%v!$H$1:$I$1", sheetName),
			},
		},
	}))

	die(file.AddChart(sheetName, "L17", &excelize.Chart{
		Type: excelize.BarStacked,
		Title: []excelize.RichTextRun{
			{
				Text: "Pass / Fail Per Package",
			},
		},
		Legend: excelize.ChartLegend{
			// Position:      "none",
			// ShowLegendKey: false,
		},
		Series: []excelize.ChartSeries{
			{
				Name:       "Pass",
				Values:     fmt.Sprintf("%v!$H$5:$H$%v", sheetName, row),
				Categories: fmt.Sprintf("%v!$G$5:$G$%v", sheetName, row),
				Fill: excelize.Fill{
					Color: []string{"00b33c"},
				},
			},
			{
				Name:       "Fail",
				Values:     fmt.Sprintf("%v!$I$5:$I$%v", sheetName, row),
				Categories: fmt.Sprintf("%v!$G$5:$G$%v", sheetName, row),
				Fill: excelize.Fill{
					Color: []string{"e60000"},
				},
			},
		},
	}))

	if err := file.SaveAs(filename); err != nil {
		fmt.Println(err)
	}
}

func die(err error) {
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
func die2(ignore any, err error) {
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
