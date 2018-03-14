package main

import (
	"archive/tar"
	"bufio"
	"flag"
	"fmt"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

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

func getNameFromFilepath(thisFile string, sunday time.Time) string {
	name := archiveName
	if archiveName != "" {
		timestamp := utils.TsRegex.FindString(thisFile)
		baseFile := path.Base(thisFile)
		ext := path.Ext(baseFile)
		filename := strings.TrimSuffix(baseFile, ext)
		name = strings.Replace(filename, timestamp, "", 1)
	}
	datedArchive := sunday.Format(utils.ArchiveForm)
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

	t, err := utils.GetTimeFromFileTimestamp(filePath)
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
		utils.EmitPath(absPath)
	} else {
		utils.EmitPath(filePath)
	}
	return nil
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("\tarchive files from directory: ")
	fmt.Printf("\t\t %s -source <source> -output <output>\n", os.Args[0])
	fmt.Println()
	fmt.Println("flags: ")
	fmt.Println()
	fmt.Println("\t-output: set the <destination> directory (default=.)")
	fmt.Println("\t-source: set the <source> directory (optional, default=stdin)")
	fmt.Println("\t-name: set the name prefix of the output tarfile <name>2006-01-02.tar (default=guess)")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
	fmt.Println("writes paths to resulting files to stdout")
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
}
