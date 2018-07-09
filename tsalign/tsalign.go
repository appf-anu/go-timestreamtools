package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	errLog                            *log.Logger
	interval                          time.Duration
	rootDir, outputDir, infmt, outfmt string
)

func alignedFilename(img utils.Image) (string, error) {

	aligned := img.Timestamp.Truncate(interval)

	targetFilename := strings.Replace(img.Path, img.Timestamp.Format(utils.TsForm), aligned.Format(utils.TsForm), 1)
	// make sure that if its already formatted as a timestream that we reformat the timestream structure.
	targetFilename = strings.Replace(targetFilename, img.Timestamp.Format(utils.DefaultTsDirectoryStructure), aligned.Format(utils.DefaultTsDirectoryStructure), 1)

	return path.Join(outputDir, targetFilename), nil
}

func moveOrRename(img utils.Image, dest string) error {
	// rename/copy+del if del is true otherwise moveFilebyCopy to not del.
	var err error

	if len(img.Data) != 0 {
		err = utils.WriteImageToFile(img, dest)
	} else {
		if err = utils.MoveFilebyCopy(img.Path, dest); err != nil {
			errLog.Printf("[move] %s", err)
			return nil
		}
	}

	return err
}

func visitWalk(filePath string, info os.FileInfo, _ error) error {
	// skip directories
	if info.IsDir() {
		return nil
	}
	img, err := utils.LoadImage(filePath)
	img.OriginalPath = filePath
	if err != nil {
		errLog.Printf("[load] %s", err)
	}

	return visit(img)
}


func visit(image utils.Image) error {
	// parse the new filepath

	newPath, err := alignedFilename(image)
	if err != nil {
		errLog.Printf("[parse] %s", err)
		return nil
	}

	newPath = filepath.Join(outputDir, filepath.Base(newPath))

	if _, err := os.Stat(newPath); err == nil {
		// skip existing.
		errLog.Printf("[skipped] %s", image.Path)
		return nil
	}

	// make directories
	err = os.MkdirAll(path.Dir(newPath), 0755)
	if err != nil {
		errLog.Printf("[mkdir] %s", err)
		return nil
	}

	absSrc, _ := filepath.Abs(image.Path)
	absDest, _ := filepath.Abs(newPath)
	if absSrc == absDest {
		errLog.Printf("[dupe] %s", absDest)
		image.Path = absDest
		utils.Emit(image, outfmt)
		return nil
	}

	if err := moveOrRename(image, absDest); err != nil {
		errLog.Printf("[move] %s", err)
		return nil
	}

	image.Path = absDest
	utils.Emit(image, outfmt)
	return err
}

var usage = func() {
	use := `
usage of %s:
flags:
	-name: renames the prefix fo the target files
	-output: set the <destination> directory (set to "tmp" to use and output a temporary dir)
	-outfmt: output format (choices: json,msgpack,path default=path)
	-infmt: input format (choices: json,msgpack,path default=path)
	-source: set the <source> directory (optional, default=stdin)
	-interval: set the interval to align to (optional, default=5m)

examples:
	align images in place:
		%s -source <source> -output <source>
	copy aligned to <destination>
		%s -source <source> -output=<destination>

will only align down, if an image is at 10:03 (5m interval) it will align to 10:00
chronologically earlier images will be kept
ie. at 5m interval, an image at 10:03 will overwrite an image at 10:02

`
	fmt.Printf(use, os.Args[0], os.Args[0], os.Args[0])
}

func init() {
	errLog = log.New(os.Stderr, "[tsalign] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset

	flag.DurationVar(&interval, "interval", time.Minute*5, "interval to align to.")
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", "", "output directory")
	flag.StringVar(&outfmt, "outfmt", "path", "output format")
	flag.StringVar(&infmt, "infmt", "path", "input format")

	// parse the leading argument with normal flag.Parse
	flag.Parse()

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
	if outputDir == "tmp" {
		tmpDir, err := ioutil.TempDir("", "tsalign-")
		if err != nil {
			panic(err)
		}
		defer utils.EmitCleanup(tmpDir, outfmt)

		outputDir = tmpDir
	}

	os.MkdirAll(outputDir, 0755)
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
