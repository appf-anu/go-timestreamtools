package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bcampbell/fuzzytime"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	errLog                 *log.Logger
	rootDir, outfmt, infmt string
	start, end             time.Time
	startTod, endTod       time.Time
)

func inTimeSpan(check time.Time) bool {
	// from: https://stackoverflow.com/questions/20924303/date-time-comparison-in-golang
	return check.After(start) && check.Before(end)
}

func inTimeOfDay(t time.Time) bool {
	st := time.Date(t.Year(), t.Month(), t.Day(), startTod.Hour(), startTod.Minute(), startTod.Second(), startTod.Nanosecond(), t.Location())
	en := time.Date(t.Year(), t.Month(), t.Day(), endTod.Hour(), endTod.Minute(), endTod.Second(), endTod.Nanosecond(), t.Location())
	return t.After(st) && t.Before(en) || t == en || t == st
}

func checkFilePath(img utils.Image) (bool, error) {
	return inTimeSpan(img.Timestamp) && inTimeOfDay(img.Timestamp), nil
}

func visitWalk(filePath string, info os.FileInfo, _ error) error {
	// skip directories
	if info.IsDir() {
		return nil
	}

	image, err := utils.LoadImage(filePath)
	image.OriginalPath = filePath
	if err != nil {
		errLog.Printf("[load] %s", err)
	}

	return visit(image)
}


func visit(img utils.Image) error {

	if ok, err := checkFilePath(img); ok {
		utils.Emit(img, outfmt)
	} else if err != nil {
		errLog.Printf("[check] %s", err)
	}

	return nil
}

var usage = func() {
	use := `
usage of %s:
flags:
	-start: the start datetime (default=1970-01-01 00:00)
	-end: the end datetime (default=now)
	-starttod: the start time of day, default 00:00:00
	-endtod: the end time of day, default 23:59:59
	-source: set the <source> directory (optional, default=stdin)
	-outfmt: output format (choices: json,msgpack,path default=path)
	-infmt: input format (choices: json,msgpack,path default=path)


examples:
	filter from 11 June 1996 until now with source:
		%s -source <source> -start 1996-06-11
	filter from 11 June 1996 to 10 December 1996 from stdin:
		%s -start 1996-06-11 -end 1996-12-10

dates are assumed to be DMY or YMD not MDY
tsselect is NON DESTRUCTIVE, and doesnt copy/move files, it only filters
`
	fmt.Printf(use, os.Args[0])
}

func parseDateTime(tString string, t *time.Time, defaultValue time.Time) error {
	ctx := fuzzytime.Context{
		DateResolver: fuzzytime.DMYResolver,
		TZResolver:   fuzzytime.DefaultTZResolver("UTC"),
	}
	datetimeValue, _, err := ctx.Extract(tString)
	if err != nil {
		errLog.Printf("[time] couldn't extract datetime: %s", err)
	}

	if datetimeValue.Empty() {
		*t = defaultValue
	} else {
		datetimeValue.Time.SetHour(datetimeValue.Time.Hour())
		datetimeValue.Time.SetMinute(datetimeValue.Time.Minute())
		datetimeValue.Time.SetSecond(datetimeValue.Time.Second())
		datetimeValue.Time.SetTZOffset(datetimeValue.Time.TZOffset())
		*t, err = time.Parse(time.RFC3339, datetimeValue.ISOFormat())
		if err != nil {
			return err
		}
	}
	return nil
}

func parseTime(tString string, t *time.Time, defaultValue time.Time) error {
	ctx := fuzzytime.Context{
		DateResolver: fuzzytime.DMYResolver,
		TZResolver:   fuzzytime.DefaultTZResolver("UTC"),
	}
	datetimeValue, _, err := ctx.Extract(tString)
	if err != nil {
		errLog.Printf("[time] couldn't extract datetime: %s", err)
	}

	if datetimeValue.Empty() {
		*t = defaultValue
	} else {
		datetimeValue.Time.SetHour(datetimeValue.Time.Hour())
		datetimeValue.Time.SetMinute(datetimeValue.Time.Minute())
		datetimeValue.Time.SetSecond(datetimeValue.Time.Second())
		datetimeValue.Time.SetTZOffset(datetimeValue.Time.TZOffset())
		*t, err = time.Parse("T15:04:05Z07:00", datetimeValue.ISOFormat())
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	errLog = log.New(os.Stderr, "[tsselect] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset

	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outfmt, "outfmt", "path", "output format")
	flag.StringVar(&infmt, "infmt", "path", "input format")

	startString := flag.String("start", "", "start datetime")
	endString := flag.String("end", "", "end datetime")
	startTodString := flag.String("starttod", "", "start time of day")
	endTodString := flag.String("endtod", "", "end time of day")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	defaultStart, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	defaultEnd, _ := time.Parse(utils.TsForm, time.Now().Format(utils.TsForm))
	defaultStartTod, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	defaultEndTod, _ := time.Parse(time.RFC3339, "1970-01-01T23:59:59Z")

	err := parseDateTime(*startString, &start, defaultStart)
	if err != nil {
		panic(err)
	}

	err = parseDateTime(*endString, &end, defaultEnd)
	if err != nil {
		panic(err)
	}

	err = parseTime(*startTodString, &startTod, defaultStartTod)
	if err != nil {
		panic(err)
	}

	err = parseTime(*endTodString, &endTod, defaultEndTod)
	if err != nil {
		panic(err)
	}

	// verify that root exists
	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				errLog.Printf("[path] <source> %s does not exist.", rootDir)
				os.Exit(1)
			}
		}
	}
}

func main() {
	if rootDir != "" {
		if err := filepath.Walk(rootDir, visitWalk); err != nil {
			errLog.Printf("[walk] %s", err)
		}
	} else {
		if infmt == "path" {
			// start scanner and wait for stdin
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				text := strings.Replace(scanner.Text(), "\n", "", -1)
				if strings.HasPrefix(text, "[") {
					errLog.Printf("[stdin] %s", text)
					continue
				} else if strings.HasPrefix(text, "#-") {
					// was signalled deletion of previous tmpdir, wait until finished
					defer os.RemoveAll(strings.TrimPrefix(text, "#-"))
				} else {
					img, err := utils.LoadImage(text)
					if err != nil {
						errLog.Printf("[load] %s", err)
					}
					visit(img)
				}
				data := strings.Replace(scanner.Text(), "\n", "", -1)
				if strings.HasPrefix(data, "[") {
					errLog.Printf("[stdin] %s", data)
					continue
				} else {
					img, err := utils.LoadImage(data)
					if err != nil {
						errLog.Printf("[load] %s", err)
					}
					visit(img)
				}
			}

		} else {
			//data := scanner.Bytes()
			//img := utils.Image{}
			//err := json.Unmarshal(data, &img)
			//if err != nil {
			//
			//	errLog.Printf("[json] %s", err)
			//	continue
			//}

			// clean up...
			//t := utils.TempDir{}
			//if err := json.Unmarshal(data, &t); err == nil{
			//	defer fmt.Printf("Removing %s\n", t.Path)
			//	defer os.RemoveAll(t.Path)
			//}
			//continue

			utils.Handle(visit, os.RemoveAll, infmt)
		}
	}
}
