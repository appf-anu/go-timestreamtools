package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/ugorji/go/codec"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	// ArchiveForm is the form that tar files should take (YYYY-MM-DD)
	ArchiveForm = "%s~2006-01-02.tar"
	// DefaultTsDirectoryStructure is the default directory structure for timestreams
	DefaultTsDirectoryStructure = "2006/2006_01/2006_01_02/2006_01_02_15/"
	// TsForm is the timestamp form for individual files.
	TsForm         = "2006_01_02_15_04_05"
	dumbExifForm   = "2006:01:02 15:04:05"
	tsRegexPattern = "[0-9][0-9][0-9][0-9]_[0-1][0-9]_[0-3][0-9]_[0-2][0-9]_[0-5][0-9]_[0-5][0-9]"
)

var (
	errLog        *log.Logger
	jsonEncoder   *json.Encoder
	jsonDecoder   *json.Decoder
	mh      codec.MsgpackHandle
	//jh      codec.JsonHandle
	msgpackDecoder *codec.Decoder
	msgpackEncoder *codec.Encoder
)

const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
)

type Image struct {
	Path            string    `json:"path"`
	OriginalPath    string    `json:"originalPath"`
	Timestamp       time.Time `json:"timestamp"`
	ExifTimestamp   time.Time `json:"exifTimestamp"`
	ExifBytes       []byte    `json:"-"`
	Data            []byte    `json:"-"`
	CmdList         []string  `json:"cmdList"`
	TempCleanupPath string    `json:"temp_cleanup_path,omitempty"`
}

type TempDir struct {
	Path string
}

func Emit(img Image, outfmt string) error {
	switch  outfmt {
	case "path":
		_, err := fmt.Fprintln(os.Stdout, img.Path)
		return err
	case "json":
		return jsonEncoder.Encode(img)
	case "msgpack":
		return msgpackEncoder.Encode(img)
	}
	return jsonEncoder.Encode(img)
}

//func EmitJson(img Image) error {
//
//	//jsonStr, err := json.Marshal(img)
//	//if err != nil {
//	//	return 0, err
//	//}
//	//return fmt.Fprintln(os.Stdout, string(jsonStr))
//	// this should be faster as it emits as a stream and leaves stdout open rather than syncing
//	return jsonEncoder.Encode(img)
//}

func EmitCleanup(tmpDir, outfmt string) error{
	// pass delete dir onto next step once finished
	switch outfmt {
	case "path":
		_, err := fmt.Fprintln(os.Stdout, "#-"+tmpDir)
		return err
	case "json":
		return jsonEncoder.Encode(Image{TempCleanupPath: tmpDir})
	case "msgpack":
		return msgpackEncoder.Encode(Image{TempCleanupPath: tmpDir})
	}
	_, err := fmt.Fprintln(os.Stdout, "#-"+tmpDir)
	return err
}

type HandleImageFn func(img Image) error
type HandleTempFn func(path string) error

func HandleJSON(handleImageFn HandleImageFn, cleanupFn HandleTempFn) error {

	for {
		img := Image{}
		err := jsonDecoder.Decode(&img)
		if img.TempCleanupPath != "" {
			defer cleanupFn(img.TempCleanupPath)
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			errLog.Println(err)
		}
		err = handleImageFn(img)
		if err != nil {
			return err
		}
	}
	return nil
}

func Handle(handleImageFn HandleImageFn, cleanupFn HandleTempFn, infmt string) error {

	for {
		img := Image{}
		var err error
		switch infmt {
		case "json":
			err = jsonDecoder.Decode(&img)
		case "msgpack":
			err = msgpackDecoder.Decode(&img)
		}

		if img.TempCleanupPath != "" {
			defer cleanupFn(img.TempCleanupPath)
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			errLog.Println(err)
			continue
		}
		err = handleImageFn(img)
		if err != nil {
			return err
		}
	}
	return nil
}

func getDtFromExif(exifData *exif.Exif) (datetime time.Time, err error) {
	// get the exif datetime
	dt, err := exifData.Get(exif.DateTime)
	if err != nil {
		return
	}
	// get string value
	datetimeStr, err := dt.StringVal()
	if err != nil {
		return
	}
	// parse string value
	if datetime, err = ParseExifDatetime(datetimeStr); err != nil {
		return
	}
	return
}

func LoadImage(imgPath string) (img Image, err error) {
	// is dot?
	if strings.HasPrefix(filepath.Base(img.Path), ".") {
		err = fmt.Errorf("[path] ignore dotfile: " + img.Path)
	}
	// stat
	finfo, err := os.Stat(imgPath)
	if err != nil {
		return
	}
	// is dir?
	if finfo.IsDir() {
		err = fmt.Errorf("[stat] is a dir: " + imgPath)
		return
	}

	// open image file
	file, err := os.Open(imgPath)
	if err != nil {
		return
	}
	// close file once loaded
	defer file.Close()
	img.Path, err = filepath.Abs(imgPath)
	if err != nil {
		img.Path = imgPath
	}

	// decode the exif data
	exifData, err := exif.Decode(file)
	if err == nil {
		// only do this if we read the exif ok
		img.ExifBytes = exifData.Raw

		if exifTimestamp, exifErr := getDtFromExif(exifData); exifErr == nil {
			// only do this if we could get the exif datetime
			img.ExifTimestamp = exifTimestamp
		}
	}

	if timestamp, err := GetTimeFromFileTimestamp(imgPath); err == nil {
		img.Timestamp = timestamp
	}

	// make sure we seek back
	file.Seek(0, io.SeekStart)

	if len(img.Data) != 0 {
		// read the image bytes into the img.Data
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(file)
		if err != nil {
			return
		}
		img.Data = buf.Bytes()
		buf.Reset()
	}

	return
}

func WriteImageToFile(img Image, destPath string) (err error) {
	if len(img.Data) == 0 {
		err = fmt.Errorf("[write] image has no data")
	}

	os.MkdirAll(path.Dir(destPath), 0770)
	err = ioutil.WriteFile(destPath, img.Data, 0770)
	return err
}

// TsRegex is a regexp to find a timestamp within a filename
var /* const */ TsRegex = regexp.MustCompile(tsRegexPattern)

// ParseExifDatetime parses a datetime string from the old dumb exif way to a time.Time{}
func ParseExifDatetime(datetimeString string) (time.Time, error) {
	thisTime, err := time.Parse(dumbExifForm, datetimeString)
	if err != nil {
		return time.Time{}, err
	}
	return thisTime, nil
}

type exifFromJSON struct {
	DateTime          string
	DateTimeOriginal  string
	DateTimeDigitized string
}

// GetTimeFromExif gets a time.Time from either the exif in an image, or the exif json for that image
func GetTimeFromExif(thisFile string) (datetime time.Time, err error) {

	var datetimeString string
	if _, ferr := os.Stat(thisFile + ".json"); ferr == nil {
		eData := exifFromJSON{}
		//	do something with the json.

		byt, err := ioutil.ReadFile(thisFile + ".json")
		if err != nil {
			return time.Time{}, err
		}
		if err := json.Unmarshal(byt, &eData); err != nil {
			return time.Time{}, err
		}
		datetimeString = eData.DateTime

	} else {
		fileHandler, err := os.Open(thisFile)
		if err != nil {
			// file wouldnt open
			return time.Time{}, err
		}
		exifData, err := exif.Decode(fileHandler)
		if err != nil {
			// exif wouldnt decode
			return time.Time{}, fmt.Errorf("[exif] couldn't decode exif from image %s", err)
		}
		dt, err := exifData.Get(exif.DateTime) // normally, don't ignore errors!
		if err != nil {
			// couldnt get DateTime from exifex
			return time.Time{}, err
		}
		datetimeString, err = dt.StringVal()
		if err != nil {
			// couldnt get
			return time.Time{}, err
		}
	}
	if datetime, err = ParseExifDatetime(datetimeString); err != nil {
		return
	}
	return
}

// GetTimeFromFileTimestamp gets a time.Time from the timestamp of an image
func GetTimeFromFileTimestamp(thisFile string) (time.Time, error) {
	timestamp := TsRegex.FindString(thisFile)
	if len(timestamp) < 1 {
		// no timestamp found in filename
		return time.Time{}, fmt.Errorf("failed regex timestamp from filename %s", thisFile)
	}

	t, err := time.Parse(TsForm, timestamp)
	if err != nil {
		// parse error
		return time.Time{}, err
	}
	return t, nil
}

// MoveFilebyCopy copiers a file
func MoveFilebyCopy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	fileMode := os.FileMode(OS_USER_RW | OS_GROUP_RW)
	d.Chmod(fileMode)

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	return d.Close()
}

func init() {
	jsonEncoder = json.NewEncoder(os.Stdout)
	//jsonDecoder = json.NewDecoder(os.Stdin)
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	jsonEncoder = json.NewEncoder(os.Stdout)
		//encoder = codec.NewEncoder(os.Stdout, &jh)
	//case "msgpack":
	msgpackEncoder = codec.NewEncoder(os.Stdout, &mh)
	//}

	jsonDecoder= json.NewDecoder(os.Stdin)
	//decoder= codec.NewDecoder(os.Stdin, &jh)
	msgpackDecoder = codec.NewDecoder(os.Stdin, &mh)
	errLog = log.New(os.Stderr, "[util] ", log.Ldate|log.Ltime|log.Lshortfile)
}
