package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/bcampbell/fuzzytime"
	"github.com/rwcarlsen/goexif/exif"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	tsForm         = "2006_01_02_15_04_05"
	dumbExifForm   = "2006:01:02 15:04:05"
	tsRegexPattern = "[0-9][0-9][0-9][0-9]_[0-1][0-9]_[0-3][0-9]_[0-2][0-9]_[0-5][0-9]_[0-5][0-9]"
)

var (
	errLog       *log.Logger
	rootDir      string
	start, end   time.Time
	datetimeFunc datetimeFunction
)

var /* const */ tsRegex = regexp.MustCompile(tsRegexPattern)

func emitPath(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stdout, a...)
}

type datetimeFunction func(string) (time.Time, error)

func parseExifDatetime(datetimeString string) (time.Time, error) {
	thisTime, err := time.Parse(dumbExifForm, datetimeString)
	if err != nil {
		return time.Time{}, err
	}
	return thisTime, nil
}

type exifFromJSON struct {
	DateTime          string
	DateTimeOriginal  string
	DateTimeDigitized string
}

func getTimeFromExif(thisFile string) (datetime time.Time, err error) {

	var datetimeString string
	if _, ferr := os.Stat(thisFile + ".json"); ferr == nil {
		eData := exifFromJSON{}
		//	do something with the json.

		byt, err := ioutil.ReadFile(thisFile + ".json")
		if err != nil {
			errLog.Printf("[json] cant read file %s", err)
		}
		if err := json.Unmarshal(byt, &eData); err != nil {
			errLog.Printf("[json] can't unmarshal %s", err)
		}

		datetimeString = eData.DateTime

	} else {
		fileHandler, err := os.Open(thisFile)
		if err != nil {

			// file wouldnt open
			return time.Time{}, err
		}
		exifData, err := exif.Decode(fileHandler)
		if err != nil {
			// exif wouldnt decode
			return time.Time{}, fmt.Errorf("[exif] couldn't decode exif from image %s", err)
		}
		dt, err := exifData.Get(exif.DateTime) // normally, don't ignore errors!
		if err != nil {
			// couldnt get DateTime from exifex
			return time.Time{}, err
		}
		datetimeString, err = dt.StringVal()
		if err != nil {
			// couldnt get
			return time.Time{}, err
		}
	}
	if datetime, err = parseExifDatetime(datetimeString); err != nil {
		errLog.Printf("[parse] parse datetime %s", err)
	}
	return
}

func getTimeFromFileTimestamp(thisFile string) (time.Time, error) {
	timestamp := tsRegex.FindString(thisFile)
	if len(timestamp) < 1 {
		// no timestamp found in filename
		return time.Time{}, errors.New("failed regex timestamp from filename")
	}

	t, err := time.Parse(tsForm, timestamp)
	if err != nil {
		// parse error
		return time.Time{}, err
	}
	return t, nil
}

func inTimeSpan(check time.Time) bool {
	// from: https://stackoverflow.com/questions/20924303/date-time-comparison-in-golang
	return check.After(start) && check.Before(end)
}

func checkFilePath(thisFile string) (bool, error) {
	thisTime, err := datetimeFunc(thisFile)
	if err != nil {
		return false, err
	}
	return inTimeSpan(thisTime), nil
}

func visit(filePath string, info os.FileInfo, _ error) error {
	// skip directories
	if info.IsDir() {
		return nil
	}

	if ok, err := checkFilePath(filePath); ok {
		emitPath(filePath)
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
	fmt.Println("\t-exif: uses exif data to get time instead of the file timestamp")
	fmt.Println("\t-source: set the <source> directory (optional, default=stdin)")
  fmt.Println()
  fmt.Println("dates are assumed to be DMY or YMD not MDY")
  fmt.Println()
	fmt.Println("reads filepaths from stdin")
  fmt.Println("writes paths to resulting files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")

}

func init() {
	errLog = log.New(os.Stderr, "[tsselect] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset

	flag.StringVar(&rootDir, "source", "", "source directory")
	startString := flag.String("start", "", "start datetime")
	endString := flag.String("end", "", "end datetime")
	useExif := flag.Bool("exif", false, "use exif instead of timestamps in filenames")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	if *useExif {
		datetimeFunc = getTimeFromExif
	} else {
		datetimeFunc = getTimeFromFileTimestamp
	}

	ctx := fuzzytime.Context{
		fuzzytime.DMYResolver,
		fuzzytime.DefaultTZResolver("UTC"),
	}
	startDatetime, _, err := ctx.Extract(*startString)
	if err != nil {
		errLog.Printf("[time] couldn't extract start datetime: %s", err)
	}

	if startDatetime.Empty() {
		start, _ = time.Parse(time.RFC3339, "1970-01-01T00:00:00Z00:00")
	} else {
		startDatetime.Time.SetHour(startDatetime.Time.Hour())
		startDatetime.Time.SetMinute(startDatetime.Time.Minute())
		startDatetime.Time.SetSecond(startDatetime.Time.Second())
		startDatetime.Time.SetTZOffset(startDatetime.Time.TZOffset())
		start, _ = time.Parse(time.RFC3339, startDatetime.ISOFormat())
	}
	endDatetime, _, err := ctx.Extract(*endString)
	if err != nil {
		errLog.Printf("[time] couldn't extract end datetime: %s", err)
	}

	if endDatetime.Empty() {
		end = time.Now()
	} else {
		endDatetime.Time.SetHour(endDatetime.Time.Hour())
		endDatetime.Time.SetMinute(endDatetime.Time.Minute())
		endDatetime.Time.SetSecond(endDatetime.Time.Second())
		endDatetime.Time.SetTZOffset(endDatetime.Time.TZOffset())
		end, _ = time.Parse(time.RFC3339, endDatetime.ISOFormat())
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

	//c := make(chan error)
	//go func() {
	//	c <- filepath.Walk(rootDir, visit)
	//}()
	//
	//if err := <-c; err != nil {
	//	fmt.Println(err)
	//}
}
