package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/oliamb/cutter"
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
	"sync"
)

var (
	rootDir, outputDir, targetExtension string
	corner1, corner2, gridxy, chunkSize image.Point
	stdin, center                       bool
	imageEncoder                        imgio.Encoder
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

func MinMax(a, b int) (min, max int) {
	if a < b {
		min = a
		max = b
	} else {
		min = b
		max = a
	}
	return
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

func writeExifJson(sourcePath, destPath string) {
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

func cropImage(sourcePath, destPath string) (err error) {
	img, err := imgio.Open(sourcePath)
	if err != nil {
		return
	}
	var cropImage image.Image
	if center {
		cropImage, err = cutter.Crop(img, cutter.Config{
			Width:  corner1.X,
			Height: corner1.Y,
			Mode:   cutter.Centered,
		})
	} else {
		cropImage, err = cutter.Crop(img, cutter.Config{
			Width:  corner2.X - corner1.X,
			Height: corner2.Y - corner1.Y,
			Anchor: image.Point{corner1.X, corner1.Y},
			Mode:   cutter.TopLeft, // optional, default value
		})
	}
	var wg sync.WaitGroup
	wg.Add(gridxy.X*gridxy.Y)

	for xPos := 0; xPos < gridxy.X; xPos++ {
		for yPos := 0; yPos < gridxy.Y; yPos++ {
			go func(xPos, yPos int){
				defer wg.Done()
				cropped, cropErr := cutter.Crop(cropImage, cutter.Config{
					Width:  chunkSize.X,
					Height: chunkSize.Y,
					Anchor: image.Point{chunkSize.X * xPos, chunkSize.Y * yPos},
					Mode:   cutter.TopLeft, // optional, default value
				})

				if cropErr != nil {
					ERRLOG("[crop] error cropping: %s", cropErr)
					return
				}

				destPos := fmt.Sprintf("%d,%d", xPos, yPos)
				destPath := fmt.Sprintf(destPath, destPos)
				os.MkdirAll(path.Dir(destPath), 0755)

				writeErr := imgio.Save(destPath, cropped, imageEncoder)
				if writeErr != nil {
					ERRLOG("[crop] error saving crop: %s", cropErr)
					return
				}
				// output the relative image path
				OUTPUT(destPath)
			}(xPos, yPos)
		}
	}
	wg.Wait()

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
	writeExifJson(filePath, path.Join(outputDir, newBase))
	newPath := path.Join(outputDir, "%s", newBase)

	// convert the image
	if err := cropImage(filePath, newPath); err != nil {
		ERRLOG("[crop] %s", err)
		return nil
	}
	return nil
}

var usage = func() {
	ERRLOG("usage of %s:", os.Args[0])
	ERRLOG("centered crop to 1920x1080")
	ERRLOG("\t%s -center -c1 1920,1080", os.Args[0])
	ERRLOG("cut out 120,10 to 400,60")
	ERRLOG("\t%s -c1 120,10 c2 400,60", os.Args[0])
	ERRLOG("centered crop to 1920x1080 and output to <destination>")
	ERRLOG("\t%s -center -c1 1920,1080 -output <destination>", os.Args[0])
	ERRLOG("")
	ERRLOG("flags:")
	pwd, _ := os.Getwd()
	ERRLOG("\t-center: center the crop, specify width,height with c1 ")
	ERRLOG("\t-c1: corner 1 (in pixels, comma separated)")
	ERRLOG("\t-c2: corner 2 (in pixels, comma separated, ignored if center is specified)")
	ERRLOG("\t-grid: split the area into this many equal crops (default=1,1)")
	ERRLOG("\t-type: set the output image type (default=jpeg)")
	ERRLOG("\t\tavailable image types:")
	ERRLOG("\t\tjpeg, png")
	ERRLOG("\t\ttiff: tiff with Deflate compression (alias for tiff-deflate)")
	ERRLOG("\t\ttiff-lzw: tiff with LZW compression")
	ERRLOG("\t\ttiff-none: tiff with no compression")
	ERRLOG("\t-output: set the <destination> directory (default=%s/<crop>)", pwd)
	ERRLOG("")
	ERRLOG("writes paths to resulting files to stdout")
	ERRLOG("reads filepaths from stdin")
	ERRLOG("will ignore any line from stdin that isnt a filepath (and only a filepath)")
}

func stringToPoint(str, sep string) (image.Point, error) {
	var err error
	ra := strings.Split(str, sep)
	if len(ra) < 2 {
		return image.Point{}, errors.New("not enough values to form point")
	}
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
	flag.Usage = usage
	// set flags for flag
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", "", "output directory")
	flag.BoolVar(&center, "center", false, "center crop")
	outputType := flag.String("type", "jpeg", "output image type")
	c1 := flag.String("c1", "0,0", "corner 1")
	c2 := flag.String("c2", "0,0", "corner 2, (ignored when center")
	grid := flag.String("grid", "1,1", "split the area into this many equal crops")

	// parse the leading argument with normal flag.Parse
	flag.Parse()

	// parse flags using a flag, ignore the first 2 (first arg is program name)
	switch *outputType {
	case "jpeg":
		imageEncoder = imgio.JPEGEncoder(95)
		targetExtension = "jpeg"
	case "tiff":
		imageEncoder = TIFFEncoder(tiff.Deflate)
		targetExtension = "tif"
	case "tiff-lzw":
		imageEncoder = TIFFEncoder(tiff.LZW)
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
	var err error
	corner1, err = stringToPoint(*c1, ",")
	if err != nil {
		ERRLOG("[flag] %s", err)
		panic(err)
	}
	corner2, err = stringToPoint(*c2, ",")
	if err != nil && !center {
		ERRLOG("[flag] %s", err)
		panic(err)
	}

	if !center {
		// sort corners for topleft and topright if not centered
		corner1.X, corner2.X = MinMax(corner1.X, corner2.X)
		corner1.Y, corner2.Y = MinMax(corner1.Y, corner2.Y)
		if corner1 == corner2 {
			ERRLOG("[flag] crop area is 0")
			os.Exit(2)
		}
	}

	gridxy, err = stringToPoint(*grid, ",")
	if err != nil {
		ERRLOG("[flag] %s", err)
		panic(err)
	}
	chunkSize = image.Point{}
	if !center {
		chunkSize.X = (corner2.X - corner1.X) / gridxy.X
		chunkSize.Y = (corner2.Y - corner1.Y) / gridxy.Y
	} else {
		chunkSize.X = corner1.X / gridxy.X
		chunkSize.Y = corner1.Y / gridxy.Y
	}

	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				ERRLOG("[path] <source> %s does not exist.", rootDir)
				os.Exit(1)
			}
		}
	}

	if outputDir == "" {
		cropFmtCenter := fmt.Sprintf("%d,%d", corner1.X, corner1.Y)
		cropFmt := fmt.Sprintf("%d,%d-%d,%d", corner1.X, corner1.Y, corner2.X, corner2.Y)
		if center {
			outputDir = path.Join(".", cropFmtCenter)
		} else {
			outputDir = path.Join(".", cropFmt)
		}
	}
	os.MkdirAll(outputDir, 0755)
	stdin = rootDir == ""

}

func main() {
	if !stdin {
		c := make(chan error)
		go func() {
			c <- filepath.Walk(rootDir, visit)
		}()

		if err := <-c; err != nil {
			ERRLOG("[goroutine] %s", err)
		}
		return
		//if err := filepath.Walk(rootDir, visit); err != nil {
		//	ERRLOG("[walk] %s", err)
		//}
		//return
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

}
