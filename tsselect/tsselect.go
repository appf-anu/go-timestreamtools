package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bcampbell/fuzzytime"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	errLog           *log.Logger
	rootDir          string
	start, end       time.Time
	datetimeFunc     datetimeFunction
	startTod, endTod time.Time
)

type datetimeFunction func(string) (time.Time, error)

func inTimeSpan(check time.Time) bool {
	// from: https://stackoverflow.com/questions/20924303/date-time-comparison-in-golang
	return check.After(start) && check.Before(end)
}

func inTimeOfDay(t time.Time) bool {
	st := time.Date(t.Year(), t.Month(), t.Day(), startTod.Hour(), startTod.Minute(), startTod.Second(), startTod.Nanosecond(), t.Location())
	en := time.Date(t.Year(), t.Month(), t.Day(), endTod.Hour(), endTod.Minute(), endTod.Second(), endTod.Nanosecond(), t.Location())
	return t.After(st) && t.Before(en) || t == en || t == st
}

func checkFilePath(thisFile string) (bool, error) {
	thisTime, err := datetimeFunc(thisFile)
	if err != nil {
		return false, err
	}
	return inTimeSpan(thisTime) && inTimeOfDay(thisTime), nil
}

func visit(filePath string, info os.FileInfo, _ error) error {
	// skip directories
	if info.IsDir() {
		return nil
	}
	if path.Ext(filePath) == ".json" {
		return nil
	}

	if strings.HasPrefix(filepath.Base(filePath), ".") {
		return nil
	}

	if ok, err := checkFilePath(filePath); ok {
		utils.EmitPath(filePath)
	} else if err != nil {
		errLog.Printf("[check] %s", err)
	}

	return nil
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("\tfilter from 11 June 1996 until now with source:")
	fmt.Printf("\t\t %s -source <source> -start 1996-06-11\n", os.Args[0])
	fmt.Println("\tfilter from 11 June 1996 to 10 December 1996 from stdin:")

	fmt.Printf("\t\t %s -start 1996-06-11 -end 1996-12-10\n", os.Args[0])
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println()
	fmt.Println("\t-start: the start datetime (default=1970-01-01 00:00)")
	fmt.Println("\t-end: the end datetime (default=now)")
	fmt.Println("\t-starttod: the start time of day, default 00:00:00")
	fmt.Println("\t-endtod: the end time of day, default 23:59:59")
	fmt.Println("\t-exif: uses exif data to get time instead of the file timestamp")
	fmt.Println("\t-source: set the <source> directory (optional, default=stdin)")
	fmt.Println()
	fmt.Println("dates are assumed to be DMY or YMD not MDY")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
	fmt.Println("writes paths to selected files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")
	fmt.Println("tsselect is NON DESTRUCTIVE, and doesnt copy/move files, it only filters")
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
		fmt.Println(defaultValue)
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
	startString := flag.String("start", "", "start datetime")
	endString := flag.String("end", "", "end datetime")
	startTodString := flag.String("starttod", "", "start time of day")
	endTodString := flag.String("endtod", "", "end time of day")
	useExif := flag.Bool("exif", false, "use exif instead of timestamps in filenames")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	if *useExif {
		datetimeFunc = utils.GetTimeFromExif
	} else {
		datetimeFunc = utils.GetTimeFromFileTimestamp
	}

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
	fmt.Println(defaultEndTod)

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
		if err := filepath.Walk(rootDir, visit); err != nil {
			errLog.Printf("[walk] %s", err)
		}
	} else {
		// start scanner and wait for stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			text := strings.Replace(scanner.Text(), "\n", "", -1)
			if strings.HasPrefix(text, "[") {
				errLog.Printf("[stdin] %s", text)
				continue
			} else {
				finfo, err := os.Stat(text)
				if err != nil {
					errLog.Printf("[stat] %s", text)
					continue
				}
				visit(text, finfo, nil)
			}
		}
	}
}
