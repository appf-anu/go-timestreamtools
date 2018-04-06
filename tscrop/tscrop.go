package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"github.com/oliamb/cutter"
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
	"sync"
)

var (
	errLog                              *log.Logger
	rootDir, outputDir, targetExtension string
	corner1, corner2, gridxy, chunkSize image.Point
	stdin, center                       bool
	imageEncoder                        imgio.Encoder
)

// TIFFEncoder returns an encoder to the Tagged Image Format
func TIFFEncoder(compressionType tiff.CompressionType) imgio.Encoder {
	return func(w io.Writer, img image.Image) error {
		return tiff.Encode(w, img, &tiff.Options{Compression: compressionType})
	}
}

func minMax(a, b int) (min, max int) {
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
	wg.Add(gridxy.X * gridxy.Y)

	for xPos := 0; xPos < gridxy.X; xPos++ {
		for yPos := 0; yPos < gridxy.Y; yPos++ {
			go func(xPos, yPos int) {
				defer wg.Done()
				cropped, cropErr := cutter.Crop(cropImage, cutter.Config{
					Width:  chunkSize.X,
					Height: chunkSize.Y,
					Anchor: image.Point{chunkSize.X * xPos, chunkSize.Y * yPos},
					Mode:   cutter.TopLeft, // optional, default value
				})

				if cropErr != nil {
					errLog.Printf("[crop] error cropping: %s", cropErr)
					return
				}

				destPos := fmt.Sprintf("%d,%d", xPos, yPos)
				destPath := fmt.Sprintf(destPath, destPos)
				os.MkdirAll(path.Dir(destPath), 0755)

				writeErr := imgio.Save(destPath, cropped, imageEncoder)
				if writeErr != nil {
					errLog.Printf("[crop] error saving crop: %s", cropErr)
					return
				}
				// output the relative image path
				utils.EmitPath(destPath)
			}(xPos, yPos)
		}
	}
	wg.Wait()

	return
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
	writeExifJSON(filePath, path.Join(outputDir, newBase))
	newPath := path.Join(outputDir, "%s", newBase)

	// convert the image
	if err := cropImage(filePath, newPath); err != nil {
		errLog.Printf("[crop] %s", err)
		return nil
	}
	return nil
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("centered crop to 1920x1080")
	fmt.Printf("\t%s -center -c1 1920,1080\n", os.Args[0])
	fmt.Println("cut out 120,10 to 400,60")
	fmt.Printf("\t%s -c1 120,10 c2 400,60\n", os.Args[0])
	fmt.Println("centered crop to 1920x1080 and output to <destination>")
	fmt.Printf("\t%s -center -c1 1920,1080 -output <destination>\n", os.Args[0])
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println()
	fmt.Println("\t-center: center the crop, specify width,height with c1")
	fmt.Println("\t-c1: corner 1 (in pixels, comma separated)")
	fmt.Println("\t-c2: corner 2 (in pixels, comma separated, ignored if center is specified)")
	fmt.Println("\t-grid: split the area into this many equal crops (default=1,1)")
	fmt.Println("\t-type: set the output image type (default=jpeg)")
	fmt.Println("\t\tavailable image types:")
	fmt.Println()
	fmt.Println("\t\tjpeg, png")
	fmt.Println("\t\ttiff: tiff with Deflate compression (alias for tiff-deflate)")
	fmt.Println("\t\ttiff-none: tiff with no compression")
	fmt.Println("\t-output: set the <destination> directory (default=<cwd>/<crop>)")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
	fmt.Println("writes paths to resulting files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")
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
	errLog = log.New(os.Stderr, "[tscrop] ", log.Ldate|log.Ltime|log.Lshortfile)
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
		errLog.Printf("[flag] %s", err)
		panic(err)
	}
	corner2, err = stringToPoint(*c2, ",")
	if err != nil && !center {
		errLog.Printf("[flag] %s", err)
		panic(err)
	}

	if !center {
		// sort corners for topleft and topright if not centered
		corner1.X, corner2.X = minMax(corner1.X, corner2.X)
		corner1.Y, corner2.Y = minMax(corner1.Y, corner2.Y)
		if corner1 == corner2 {
			errLog.Printf("[flag] crop area is 0")
			os.Exit(2)
		}
	}

	gridxy, err = stringToPoint(*grid, ",")
	if err != nil {
		errLog.Printf("[flag] %s", err)
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
				errLog.Printf("[path] <source> %s does not exist.", rootDir)
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
			errLog.Printf("[goroutine] %s", err)
		}
		return
		//if err := filepath.Walk(rootDir, visit); err != nil {
		//	errLog.Printf("[walk] %s", err)
		//}
		//return
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

}
