package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	errLog             *log.Logger
	interval           time.Duration
	rootDir, outputDir string
	del, stdin         bool
	datetimeFunc       datetimeFunction
)

type datetimeFunction func(string) (time.Time, error)

func alignedFilename(thisFile string) (string, error) {
	thisTime, err := datetimeFunc(thisFile)
	if err != nil {
		return "", err
	}

	aligned := thisTime.Truncate(interval)

	if err != nil {
		return "", err
	}

	targetFilename := strings.Replace(thisFile, thisTime.Format(utils.TsForm), aligned.Format(utils.TsForm), 1)
	// make sure that if its already formatted as a timestream that we reformat the timestream structure.
	targetFilename = strings.Replace(targetFilename, thisTime.Format(utils.DefaultTsDirectoryStructure), aligned.Format(utils.DefaultTsDirectoryStructure), 1)
	if del {
		return targetFilename, nil
	}

	return path.Join(outputDir, targetFilename), nil
}

func moveOrRename(source, dest string) error {
	// rename/copy+del if del is true otherwise moveFilebyCopy to not del.
	var err error
	if del {
		err = os.Rename(source, dest)
		if err != nil {
			err = utils.MoveFilebyCopy(source, dest, del)
		}
	} else {
		err = utils.MoveFilebyCopy(source, dest, del)
	}
	if err != nil {
		errLog.Printf("[move] %s", err)
		return nil
	}
	return err
}

func visit(filePath string, info os.FileInfo, _ error) error {
	// skip directories
	if info.IsDir() {
		return nil
	}
	if path.Ext(filePath) == ".json" {
		return nil
	}

	// parse the new filepath
	newPath, err := alignedFilename(filePath)
	if err != nil {
		errLog.Printf("[parse] %s", err)
		return nil
	}
	newPath = filepath.Join(outputDir, strings.Replace(newPath, rootDir, "", 1))

	if _, err := os.Stat(newPath); err == nil {
		// skip existing.
		errLog.Printf("[skipped] %s", filePath)
		return nil
	}

	// make directories
	err = os.MkdirAll(path.Dir(newPath), 0755)
	if err != nil {
		errLog.Printf("[mkdir] %s", err)
		return nil
	}

	absSrc, _ := filepath.Abs(filePath)
	absDest, _ := filepath.Abs(newPath)
	if absSrc == absDest {
		errLog.Printf("[dupe] %s", absDest)
		return nil
	}

	err = moveOrRename(filePath, absDest)
	jsFile := filePath + ".json"
	if _, ferr := os.Stat(jsFile); ferr == nil {
		if e := moveOrRename(jsFile, absDest+".json"); e != nil {
			errLog.Println("[exif] couldn't move json exif file")
		}
	}
	utils.EmitPath(newPath)
	return err
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("\talign images in place:")
	fmt.Printf("\t\t%s -source <source> -output <source>\n", os.Args[0])
	fmt.Println("\t copy aligned to <destination>:")
	fmt.Printf("\t\t%s -source <source> -output=<destination>\n", os.Args[0])

	fmt.Println()
	fmt.Println("flags:")
	fmt.Println()
	fmt.Println("\t-name: renames the prefix fo the target files")
	fmt.Println("\t-exif: uses exif data to rename rather than file timestamp")
	fmt.Printf("\t-output: set the <destination> directory (default=<cwd>)")
	fmt.Println("\t-source: set the <source> directory (optional, default=stdin)")
	fmt.Println("\t-interval: set the interval to align to (optional, default=5m)")
	fmt.Println()
	fmt.Println("will only align down, if an image is at 10:03 (5m interval) it will align to 10:00")
	fmt.Println("chronologically earlier images will be kept")
	fmt.Println("ie. at 5m interval, an image at 10:03 will overwrite an image at 10:02")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
	fmt.Println("writes paths to resulting files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")
}

func init() {
	errLog = log.New(os.Stderr, "[tsalign] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset

	flag.DurationVar(&interval, "interval", time.Minute*5, "interval to align to.")
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", ".", "output directory")

	useExif := flag.Bool("exif", false, "use exif instead of timestamps in filenames")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	if *useExif {
		datetimeFunc = utils.GetTimeFromExif
	} else {
		datetimeFunc = utils.GetTimeFromFileTimestamp
	}

	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				errLog.Printf("[path] <source> %s does not exist.", rootDir)
				os.Exit(1)
			}
		}
	}

	os.MkdirAll(outputDir, 0755)

	stdin = rootDir == ""

	outputAbs, _ := filepath.Abs(outputDir)
	absRoot, _ := filepath.Abs(rootDir)

	// if output and source are the same then it is an in place rename.
	del = absRoot == outputAbs

}

func main() {
	if !stdin {
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
