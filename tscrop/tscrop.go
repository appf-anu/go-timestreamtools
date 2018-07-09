package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"github.com/oliamb/cutter"
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
	errLog                                             *log.Logger
	rootDir, outputDir, outfmt, infmt, targetExtension string
	corner1, corner2, gridxy, chunkSize                image.Point
	imageEncoder                                       imgio.Encoder
	center                                             bool
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

func cropImage(sourceImg utils.Image, destPath string) (err error) {

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
	wg.Add(gridxy.X * gridxy.Y) // add this number to the waitgroup so wait for all of these to finish.

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
				buf2 := new(bytes.Buffer)
				imgWriter := bufio.NewWriter(buf2)
				err = imageEncoder(imgWriter, cropped)
				if err != nil {
					return
				}

				imgWriter.Flush()

				// read the image bytes into the img.Data

				destPos := fmt.Sprintf("%d,%d", xPos, yPos)
				destPath := fmt.Sprintf(destPath, destPos)
				cImg := utils.Image{
					OriginalPath:  sourceImg.OriginalPath,
					Data:          buf2.Bytes(),
					Timestamp:     sourceImg.Timestamp,
					ExifTimestamp: sourceImg.ExifTimestamp,
					CmdList:       append(sourceImg.CmdList, strings.Join(os.Args, " ")),
				}
				buf2.Reset()

				// write image out.
				if writeErr := utils.WriteImageToFile(cImg, destPath); writeErr != nil {
					errLog.Printf("[crop] error saving crop: %s", cropErr)
					return
				}

				// output the relative image path
				utils.Emit(cImg, outfmt)

			}(xPos, yPos)
		}
	}
	wg.Wait()

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
	newPath := path.Join(outputDir, "%s", newBase)

	// convert the image
	if err := cropImage(img, newPath); err != nil {
		errLog.Printf("[crop] %s", err)
		return nil
	}
	return nil
}

var usage = func() {
	use := `
usage of %s:

flags:
	-center: center the crop, specify width,height with c1
	-c1: corner 1 (in pixels, comma separated)
	-c2: corner 2 (in pixels, comma separated, ignored if center is specified)
	-grid: split the area into this many equal crops (default=1,1)
	-type: set the output image type (default=jpeg)
	-output: set the <destination> directory (default=<cwd>/<crop>)

available image types:
	jpeg, png
	tiff: tiff with Deflate compression (alias for tiff-deflate)
	tiff-none: tiff with no compression

examples:
	centered crop to 1920x1080:
		%s -center -c1 1920,1080
	cut out 120,10 to 400,60:
		%s -c1 120,10 c2 400,60
	centered crop to 1920x1080 and output to <destination>:
		%s -center -c1 1920,1080 -output <destination>
	4x4 grid crop of centered 1920x1080:
		%s -grid 4,4  -center -c1 1920,1080 -output <destination>
`
	fmt.Printf(use, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
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
	flag.StringVar(&outfmt, "outfmt", "path", "output format")
	flag.StringVar(&infmt, "infmt", "path", "input format")
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

	if center {
		cropFmtCenter := fmt.Sprintf("%d,%d", corner1.X, corner1.Y)
		outputDir = path.Join(outputDir, cropFmtCenter)
	} else {
		cropFmt := fmt.Sprintf("%d,%d-%d,%d", corner1.X, corner1.Y, corner2.X, corner2.Y)
		outputDir = path.Join(outputDir, cropFmt)
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
