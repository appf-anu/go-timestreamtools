package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type timestampTestResult struct {
	filename, exifVersion string

	expectedTime time.Time
}

var /* const */ timeStamps = []timestampTestResult{
	{
		"BVZ-House-Picam_2016_06_08_10_10_00.jpg",
		"2016:06:08 10:10:00",
		time.Date(2016, 6, 8, 10, 10, 0, 0, time.UTC),
	},

	{
		"_1998_06_08_10_10_00.jpgBVZ-House-Picam",
		"1998:06:08 10:10:00",
		time.Date(1998, 6, 8, 10, 10, 0, 0, time.UTC),
	},
	{
		"BVZ-House-Picam2012_06_08_10_10_00",
		"2012:06:08 10:10:00",
		time.Date(2012, 6, 8, 10, 10, 0, 0, time.UTC),
	},
	{
		"BVZ-House-Picam2000_02_05_01_10_00",
		"2000:02:05 01:10:00",
		time.Date(2000, 2, 5, 1, 10, 0, 0, time.UTC),
	},
}

func TestGetTimeFromFileTimestamp(t *testing.T) {
	for _, tt := range timeStamps {
		tr, err := GetTimeFromFileTimestamp(tt.filename)
		if err != nil {
			t.Error(err)
		}
		if tr != tt.expectedTime {
			t.Errorf("GetTimeFromFileTimestamp got %s expected %s", tr.Format(TsForm), tt.expectedTime.Format(TsForm))
		}
	}
}

func TestParseExifDatetime(t *testing.T) {
	for _, tt := range timeStamps {
		tr, err := ParseExifDatetime(tt.exifVersion)
		if err != nil {
			t.Error(err)
		}
		if tr != tt.expectedTime {
			assert.EqualValues(t, tr, tt.expectedTime)
			//t.Errorf("ParseExifDatetime got %s expected %s", tr.Format(TsForm), tt.expectedTime.Format(TsForm))
		}
	}
}

type tHelper interface {
	Helper()
}

func FileNotExists(t assert.TestingT, path string, msgAndArgs ...interface{}) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		return assert.Fail(t, fmt.Sprintf("error when running os.Lstat(%q): %s", path, err), msgAndArgs...)
	}
	if info.IsDir() {
		return assert.Fail(t, fmt.Sprintf("%q is a directory", path), msgAndArgs...)
	}
	return assert.Fail(t, fmt.Sprintf("file exists %q", path), msgAndArgs...)
}

func TestMoveFilebyCopy(t *testing.T) {
	f, err := os.Create("testFile")
	if err != nil {
		t.Error(err)
	}

	testByt := []byte("testBytes blahblahblah\nblahblah")
	f.Write(testByt)
	f.Close()

	err = MoveFilebyCopy("testFile", "testFileNew", false)
	if err != nil {
		t.Error(err)
	}
	assert.FileExists(t, "testFile")
	assert.FileExists(t, "testFileNew")
	rbyt, err := ioutil.ReadFile("testFileNew")
	if err != nil {
		t.Error(err)
	}
	assert.EqualValues(t, testByt, rbyt)
	os.Remove("testFileNew")

	err = MoveFilebyCopy("testFile", "testFileNew", true)
	if err != nil {
		t.Error(err)
	}
	assert.FileExists(t, "testFileNew")

	FileNotExists(t, "testFile")
	os.Remove("testFileNew")
}
