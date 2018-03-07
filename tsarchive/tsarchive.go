package main

import (
	"archive/tar"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func emitPath(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stdout, a...)
}

const (
	archiveForm    = "%s2006-01-02.tar"
	tsForm         = "2006_01_02_15_04_05"
	tsRegexPattern = "[0-9][0-9][0-9][0-9]_[0-1][0-9]_[0-3][0-9]_[0-2][0-9]_[0-5][0-9]_[0-5][0-9]"
)

var /* const */ tsRegex = regexp.MustCompile(tsRegexPattern)

var (
	errLog                          *log.Logger
	rootDir, outputDir, archiveName string
	weeklyFileWriter                []*os.File
	weeklyTarWriters                map[time.Time]*tar.Writer
)

func addFile(tw *tar.Writer, thePath string) error {
	file, err := os.Open(thePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = path.Base(thePath)
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
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

func getNameFromFilepath(thisFile string, sunday time.Time) string {
	name := archiveName
	if archiveName != "" {
		timestamp := tsRegex.FindString(thisFile)
		baseFile := path.Base(thisFile)
		ext := path.Ext(baseFile)
		filename := strings.TrimSuffix(baseFile, ext)
		name = strings.Replace(filename, timestamp, "", 1)
	}
	datedArchive := sunday.Format(archiveForm)
	return fmt.Sprintf(datedArchive, name)
}

func truncateTimeToSunday(t time.Time) (sunday time.Time) {
	return t.Truncate(time.Hour * 24 * 7)
}

func visit(filePath string, info os.FileInfo, _ error) error {
	// skip directories
	if info.IsDir() {
		return nil
	}
	ext := path.Ext(filePath)
	switch extlower := strings.ToLower(ext); extlower {
	case ".jpeg", ".jpg", ".tif", ".tiff", ".cr2":
		break
	default:
		return nil
	}

	t, err := getTimeFromFileTimestamp(filePath)
	if err != nil {
		errLog.Printf("%s", err)
		return nil
	}
	sunday := truncateTimeToSunday(t)

	if _, ok := weeklyTarWriters[sunday]; !ok {
		tarbaseName := getNameFromFilepath(filePath, sunday)
		tarPath := path.Join(outputDir, tarbaseName)
		file, err := os.Create(tarPath)
		if err != nil {
			errLog.Printf("%s", err)
			panic(err)
		}
		weeklyFileWriter = append(weeklyFileWriter, file)
		weeklyTarWriters[sunday] = tar.NewWriter(file)
		errLog.Printf("[tar] opened %s tar writer", sunday.Format("2006-01-02"))
	}

	if err := addFile(weeklyTarWriters[sunday], filePath); err != nil {
		errLog.Printf("%s", err)
		return nil
	}

	if absPath, err := filepath.Abs(filePath); err == nil {
		emitPath(absPath)
	} else {
		emitPath(filePath)
	}
	return nil
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println("\tarchive files from directory: ")
	fmt.Printf("\t\t %s -source <source> -output <output>\n", os.Args[0])

	fmt.Println("")
	fmt.Println("flags: ")
	pwd, _ := os.Getwd()
	fmt.Printf("\t-output: set the <destination> directory (default=%s)\n", pwd)
	fmt.Printf("\t-source: set the <source> directory (optional, default=stdin)\n", pwd)
	fmt.Printf("\t-name: set the name prefix of the output tarfile <name>2006-01-02.tar (default=guess)\n", pwd)
	fmt.Println("")
	fmt.Println("reads filepaths from stdin")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")
}

func init() {
	errLog = log.New(os.Stderr, "[tsarchive] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", "", "output directory")
	flag.StringVar(&archiveName, "name", "", "output directory")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	// create dirs
	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				errLog.Printf("[path] <source> %s does not exist.", rootDir)
				os.Exit(1)
			}
		}
	}

	// more create dirs
	if outputDir == "" {
		if rootDir == "" {
			outputDir, _ = os.Getwd()
		} else {
			outputDir = rootDir
		}
		errLog.Printf("[path] no <destination>, creating %s", outputDir)
	}
	if _, err := os.Stat(outputDir); err != nil {
		os.MkdirAll(outputDir, 0755)
	}
}

func main() {

	weeklyTarWriters = make(map[time.Time]*tar.Writer)
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

	for sunday, writer := range weeklyTarWriters {
		errLog.Printf("[tar] closing %s tar writer", sunday.Format("2006-01-02"))
		writer.Close()
	}

	for i := range weeklyFileWriter {
		weeklyFileWriter[i].Close()
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
