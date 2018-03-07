package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/tiff"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	errLog                              *log.Logger
	rootDir, outputDir, targetExtension string
	resolution                          image.Point
	stdin                               bool
	imageEncoder                        imgio.Encoder
)

func emitPath(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stdout, a...)
}

// TIFFEncoder returns an encoder to the Tagged Image Format
func TIFFEncoder(compressionType tiff.CompressionType) imgio.Encoder {
	return func(w io.Writer, img image.Image) error {
		return tiff.Encode(w, img, &tiff.Options{Compression: compressionType})
	}
}

func getExifData(thisFile string) ([]byte, error) {
	fileHandler, err := os.Open(thisFile)
	if err != nil {
		// file wouldnt open
		return []byte{}, err
	}

	exifData, err := exif.Decode(fileHandler)
	if err != nil {
		// exif wouldnt decode
		return []byte{}, err
	}

	jsonBytes, err := exifData.MarshalJSON()
	if err != nil {
		return []byte{}, err
	}
	return jsonBytes, nil
}

func writeExifJSON(sourcePath, destPath string) {
	exifJSON, exifJSONErr := getExifData(sourcePath)

	if exifJSONErr != nil {
		errLog.Printf("[exif] couldnt read data from %s", sourcePath)
	}
	if len(exifJSON) > 0 {
		exifJSONErr := ioutil.WriteFile(destPath+".json", exifJSON, 0644)
		if exifJSONErr != nil {
			errLog.Printf("[exif] couldnt write json %s", destPath)
		}
	}
}

func convertImage(sourcePath, destPath string) (err error) {
	writeExifJSON(sourcePath, destPath)

	img, err := imgio.Open(sourcePath)
	if err != nil {
		return
	}
	if img == nil {
		return errors.New("[imgload] Nil img wtf")
	}

	resized := transform.Resize(img, resolution.X, resolution.Y, transform.Lanczos)

	imgio.Save(destPath, resized, imageEncoder)
	return
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

	basePath := path.Base(filePath)
	// parse the new filepath
	noExtension := strings.TrimSuffix(basePath, ext)
	newBase := fmt.Sprintf("%s.%s", noExtension, targetExtension)
	newPath := path.Join(outputDir, newBase)

	// convert the image
	if err := convertImage(filePath, newPath); err != nil {
		errLog.Printf("[convert] %s", err)
		return nil
	}
	// output the relative image path
	emitPath(newPath)

	return nil
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("flags:")
  fmt.Println()
	fmt.Println("\t-res: output image resolution")
	fmt.Println("\t-output: <destination> directory (default=.)")
	fmt.Println("\t-type: output image type (default=jpeg)")
  fmt.Println()
	fmt.Println("\t\tavailable image types:")
  fmt.Println()
	fmt.Println("\t\tjpeg, png")
	fmt.Println("\t\ttiff: tiff with Deflate compression (alias for tiff-deflate)")
	fmt.Println("\t\ttiff-none: tiff with no compression")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
  fmt.Println("writes paths to resulting files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")
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
	outputType := flag.String("type", "jpeg", "output image type")
	res := flag.String("res", "", "resolution")
	flag.Parse()

	switch *outputType {
	case "jpeg":
		imageEncoder = imgio.JPEGEncoder(95)
		targetExtension = "jpeg"
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
		targetExtension = "jpeg"
	}
	if *res == "" {
		errLog.Println("[flag] no resolution specified")
		os.Exit(2)
	}

	var err error
	resolution, err = stringToPoint(*res, "x")
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
	if outputDir == "" {
		outputDir = path.Join(".", *res)
	}
	os.MkdirAll(outputDir, 0755)
	stdin = rootDir == ""

}

func main() {
	if !stdin {
		if err := filepath.Walk(rootDir, visit); err != nil {
			errLog.Printf("[walk] %s", err)
		}
		return
	}
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
				errLog.Printf("[stat] %s", err)
				continue
			}
			visit(text, finfo, nil)
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
