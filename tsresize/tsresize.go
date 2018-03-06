package main

import (
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/tiff"
	"image"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"bufio"
	"errors"
)

var (
	rootDir, outputDir, targetExtension string
	resolution                                            image.Point
	stdin                                                 bool
	imageEncoder                                          imgio.Encoder
)

func ERRLOG(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format+"\n", a...)
}

func OUTPUT(a ...interface{}) (n int, err error) {
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


func writeExifJson(sourcePath, destPath string){
	exifJson, exifJSONErr := getExifData(sourcePath)

	if exifJSONErr != nil {
		ERRLOG("[exif] couldnt read data from %s", sourcePath)
	}
	if len(exifJson) > 0 {
		exifJSONErr := ioutil.WriteFile(destPath+".json", exifJson, 0644)
		if exifJSONErr != nil {
			ERRLOG("[exif] couldnt write json %s", destPath)
		}
	}
}


func convertImage(sourcePath, destPath string) (err error) {
	writeExifJson(sourcePath, destPath)

	img, err := imgio.Open(sourcePath)
	if err != nil {
		return
	}
	if img == nil{
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
		ERRLOG("[convert] %s", err)
		return nil
	}
	// output the relative image path
	OUTPUT(newPath)

	return nil
}

var usage = func() {
	ERRLOG("usage of %s:\n", os.Args[0])

	pwd, _ := os.Getwd()
	ERRLOG("")
	ERRLOG("flags:")
	ERRLOG("\t-res: output image resolution")
	ERRLOG("\t-output: <destination> directory (default=<res>/%s)", pwd)
	ERRLOG("\t-type: output image type (default=jpeg)")
	ERRLOG("")
	ERRLOG("\t\tavailable image types:")
	ERRLOG("\t\tjpeg, png")
	ERRLOG("\t\ttiff: tiff with Deflate compression (alias for tiff-deflate)")
	ERRLOG("\t\ttiff-none: tiff with no compression")

	ERRLOG("")
	ERRLOG("writes paths to resulting files to stdout")
	ERRLOG("reads filepaths from stdin")
	ERRLOG("will ignore any line from stdin that isnt a filepath (and only a filepath)")
}

func stringToPoint(str, sep string) (image.Point, error) {
	var err error
	ra := strings.Split(str, sep)
	point := image.Point{}
	if point.X, err = strconv.Atoi(ra[0]); err !=nil{
		return image.Point{}, err
	}
	if point.Y, err = strconv.Atoi(ra[1]); err !=nil{
		return image.Point{}, err
	}

	return point, err
}

func init() {
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
	if *res == ""{
		ERRLOG("[flag] no resolution specified")
		os.Exit(2)
	}

	var err error
	resolution, err = stringToPoint(*res,"x")
	if err != nil{
		ERRLOG("[flag] %s", err)
		panic(err)
	}
	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				ERRLOG("[path] <source> %s does not exist", rootDir)
				os.Exit(1)
			}
		}
	}
	if outputDir == ""{
		outputDir = path.Join(".", *res)
	}
	os.MkdirAll(outputDir, 0755)
	stdin = rootDir == ""

}

func main() {
	if !stdin{
		if err := filepath.Walk(rootDir, visit); err != nil {
			ERRLOG("[walk] %s", err)
		}
		return
	}
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
				ERRLOG("[stat] %s", err)
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
