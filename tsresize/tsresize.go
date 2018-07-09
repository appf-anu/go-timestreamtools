package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"golang.org/x/image/tiff"
	//"github.com/mdaffin/go-telegraf"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	//"time"
	"bytes"
)

var (
	errLog                                             *log.Logger
	rootDir, outputDir, outfmt, infmt, targetExtension string
	resolution                                         image.Point
	res                                                string
	imageEncoder                                       imgio.Encoder
)

// TIFFEncoder returns an encoder to the Tagged Image Format
func TIFFEncoder(compressionType tiff.CompressionType) imgio.Encoder {
	return func(w io.Writer, img image.Image) error {
		return tiff.Encode(w, img, &tiff.Options{Compression: compressionType})
	}
}

func convertImage2(sourcePath, destPath string) (err error) {
	img, err := imgio.Open(sourcePath)
	if err != nil {
		return
	}
	if img == nil {
		return errors.New("[imgload] nil img wtf")
	}

	resized := transform.Resize(img, resolution.X, resolution.Y, transform.Lanczos)

	imgio.Save(destPath, resized, imageEncoder)
	return
}

func convertImage(sourceImg *utils.Image) (err error) {
	if len(sourceImg.Data) == 0 {
		file, err := os.Open(sourceImg.Path)
		if err != nil {
			return
		}
		// read the image bytes into the img.Data
		buf1 := new(bytes.Buffer)
		_, err = buf1.ReadFrom(file)
		if err != nil {
			return
		}

		sourceImg.Data = buf1.Bytes()
		buf1.Reset()
	}

	imgReader := bytes.NewReader(sourceImg.Data)
	img, _, err := image.Decode(imgReader)
	if err != nil {
		return
	}
	if img == nil {
		return errors.New("[imgload] nil img wtf")
	}

	resized := transform.Resize(img, resolution.X, resolution.Y, transform.Lanczos)

	buf2 := new(bytes.Buffer)
	imgWriter := bufio.NewWriter(buf2)

	err = imageEncoder(imgWriter, resized)
	if err != nil {
		return
	}

	imgWriter.Flush()
	// read the image bytes into the img.Data
	sourceImg.Data = buf2.Bytes()
	buf2.Reset()
	return
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

func visit(img utils.Image) error {

	ext := path.Ext(img.Path)
	switch extlower := strings.ToLower(ext); extlower {
	case ".jpeg", ".jpg", ".tif", ".tiff", ".cr2":
		break
	default:
		return nil
	}

	basePath := path.Base(img.Path)
	// parse the new filepath
	noExtension := strings.TrimSuffix(basePath, ext)
	newBase := fmt.Sprintf("%s.%s", noExtension, targetExtension)
	newPath := path.Join(outputDir, newBase)

	// convert the img
	if err := convertImage(&img); err != nil {
		errLog.Printf("[convert] %s", err)
		return nil
	}
	img.Path = newPath
	utils.WriteImageToFile(img, newPath)

	// output the relative img path
	utils.Emit(img, outfmt)

	return nil
}

var usage = func() {
	use := `
usage of %s:
flags:
	-res: output image resolution
	-write: output image resolution
	-output: set the <destination> directory (set to "tmp" to use and output a temporary dir
	-type: output image type (default=jpg)
	-outfmt: output format (choices: json,msgpack,path default=path)
	-infmt: input format (choices: json,msgpack,path default=path)

available image types:
	jpg, png
	tiff: tiff with Deflate compression (alias for tiff-deflate)
	tiff-none: tiff with no compression

`
	fmt.Printf(use, os.Args[0])
}

func stringToPoint(str, sep string) (image.Point, error) {
	var err error
	ra := strings.Split(str, sep)
	point := image.Point{}
	if point.X, err = strconv.Atoi(ra[0]); err != nil {
		return image.Point{}, err
	}
	if point.Y, err = strconv.Atoi(ra[1]); err != nil {
		return image.Point{}, err
	}

	return point, err
}

func init() {
	errLog = log.New(os.Stderr, "[tsresize] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flag
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", "", "output directory")
	outputType := flag.String("type", "jpg", "output image type")
	flag.StringVar(&outfmt, "outfmt", "path", "output format")
	flag.StringVar(&infmt, "infmt", "path", "input format")
	flag.StringVar(&res, "res", "", "resolution")
	flag.Parse()

	switch *outputType {
	case "jpg":
		imageEncoder = imgio.JPEGEncoder(95)
		targetExtension = "jpg"
	case "tiff":
		imageEncoder = TIFFEncoder(tiff.Deflate)
		targetExtension = "tif"
	case "tiff-deflate":
		imageEncoder = TIFFEncoder(tiff.Deflate)
		targetExtension = "tif"
	case "tiff-none":
		imageEncoder = TIFFEncoder(tiff.Uncompressed)
		targetExtension = "tif"
	case "png":
		imageEncoder = imgio.PNGEncoder()
		targetExtension = "png"
	default:
		imageEncoder = imgio.JPEGEncoder(95)
		targetExtension = "jpg"
	}
	if res == "" {
		errLog.Println("[flag] no resolution specified")
		os.Exit(2)
	}

	var err error
	resolution, err = stringToPoint(res, "x")
	if err != nil {
		errLog.Printf("[flag] %s", err)
		panic(err)
	}

	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				errLog.Printf("[path] <source> %s does not exist", rootDir)
				os.Exit(1)
			}
		}
	}
}

func main() {
	if outputDir == "tmp" {
		tmpDir, err := ioutil.TempDir("", "tsresize-")
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
