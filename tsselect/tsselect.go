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
	rootDir      string
	start, end   time.Time
	datetimeFunc datetimeFunction
)

var /* const */ tsRegex = regexp.MustCompile(tsRegexPattern)

func ERRLOG(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format+"\n", a...)
}

func OUTPUT(a ...interface{}) (n int, err error) {
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

type ExifFromJSON struct {
	DateTime          string
	DateTimeOriginal  string
	DateTimeDigitized string
}

func getTimeFromExif(thisFile string) (datetime time.Time, err error) {

	var datetimeString string
	if _, ferr := os.Stat(thisFile + ".json"); ferr == nil {
		eData := ExifFromJSON{}
		//	do something with the json.

		byt, err := ioutil.ReadFile(thisFile + ".json")
		if err != nil {
			ERRLOG("[json] cant read file %s", err)
		}
		if err := json.Unmarshal(byt, &eData); err != nil {
			ERRLOG("[json] can't unmarshal %s", err)
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
			return time.Time{}, errors.New(fmt.Sprintf("[exif] couldn't decode exif from image %s", err))
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
		ERRLOG("[parse] parse datetime %s", err)
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
		OUTPUT(filePath)
	} else if err != nil {
		ERRLOG("[check] %s", err)
	}

	return nil
}

var usage = func() {
	ERRLOG("usage of %s:", os.Args[0])
	ERRLOG("\tfilter from 11 June 1996 until now with source:")
	ERRLOG("\t\t %s -source <source> -start 1996-06-11", os.Args[0])
	ERRLOG("\tfilter from 11 June 1996 to 10 December 1996 from stdin:")
	ERRLOG("\t\t %s -start 1996-06-11 -end 1996-12-10", os.Args[0])
	ERRLOG("")
	ERRLOG("flags:")
	ERRLOG("\t-start: the start datetime (default=1970-01-01 00:00)")
	ERRLOG("\t-end: the end datetime (default=now)")
	ERRLOG("\t-exif: uses exif data to get time instead of the file timestamp")
	ERRLOG("\t-source: set the <source> directory (optional, default=stdin)")
	ERRLOG("")
	ERRLOG("reads filepaths from stdin")
	ERRLOG("will ignore any line from stdin that isnt a filepath (and only a filepath)")
	ERRLOG("dates are assumed to be DMY or YMD not MDY")

}

func init() {
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
		ERRLOG("[time] couldn't extract start datetime: %s", err)
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
		ERRLOG("[time] couldn't extract end datetime: %s", err)
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
				ERRLOG("[path] <source> %s does not exist.", rootDir)
				os.Exit(1)
			}
		}
	}

}

func main() {

	if rootDir != "" {
		if err := filepath.Walk(rootDir, visit); err != nil {
			ERRLOG("[walk] %s", err)
		}
	} else {
		// start scanner and wait for stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			text := strings.Replace(scanner.Text(), "\n", "", -1)
			if strings.HasPrefix(text, "[") {
				ERRLOG("[stdin] %s", text)
				continue
			} else {
				finfo, err := os.Stat(text)
				if err != nil {
					ERRLOG("[stat] %s", text)
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
