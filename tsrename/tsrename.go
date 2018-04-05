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
	errLog                          *log.Logger
	rootDir, outputDir, namedOutput string
	del                             bool
	datetimeFunc                    datetimeFunction
)

type datetimeFunction func(string) (time.Time, error)

func parseFilename(thisFile string) (string, error) {
	thisTime, err := datetimeFunc(thisFile)
	if err != nil {
		return "", err
	}

	formattedSubdirs := path.Dir(thisFile)

	ext := path.Ext(thisFile)
	if ext == ".jpeg"{
		ext = ".jpg"
	}
	if ext == ".tiff"{
		ext = ".tif"
	}
	targetFilename := namedOutput + "_" + thisTime.Format(utils.TsForm)+"_00" + ext

	newT := path.Join(outputDir, formattedSubdirs, targetFilename)

	return newT, nil
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
	// skip directories
	if info.IsDir() {
		return nil
	}
	if path.Ext(filePath) == ".json" {
		return nil
	}

	if strings.HasPrefix(filepath.Base(filePath), "."){
		return nil
	}

	// parse the new filepath
	newPath, err := parseFilename(filePath)
	if err != nil {
		errLog.Printf("[parse] %s", err)
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
			errLog.Printf("[exif] couldn't move json exif file")
		}
	}

	utils.EmitPath(newPath)

	return err
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("\tcopy with <name> prefix:")
	fmt.Printf("\t\t %s -source <source> -name=<name>\n", os.Args[0])
	fmt.Println("\tcopy with <name> prefix:")
	fmt.Printf("\t\t %s -source <source> -name=<name>\n", os.Args[0])
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("\t-del: removes the source files")
	fmt.Println("\t-name: renames the prefix fo the target files")
	fmt.Println("\t-exif: uses exif data to rename rather than file timestamp")
	fmt.Println("\t-output: set the <destination> directory (default=.)")
	fmt.Println("\t-source: set the <source> directory (optional, default=stdin)")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
	fmt.Println("writes paths to resulting files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")

}

func init() {
	errLog = log.New(os.Stderr, "[tsrename] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset
	flag.StringVar(&namedOutput, "name", "", "name for the stream")
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", ".", "output directory")
	flag.BoolVar(&del, "del", false, "delete source files")

	useExif := flag.Bool("exif", false, "use exif instead of timestamps in filenames")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	if *useExif {
		datetimeFunc = utils.GetTimeFromExif
	} else {
		datetimeFunc = utils.GetTimeFromFileTimestamp
	}
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
	//outputDir, _ = filepath.Abs(outputDir)
	os.MkdirAll(outputDir, 0755)
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
