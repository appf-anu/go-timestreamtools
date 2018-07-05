package main

import (
	"archive/tar"
	"bufio"
	"flag"
	"fmt"
	"github.com/borevitzlab/go-timestreamtools/utils"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	errLog                          *log.Logger
	rootDir, outputDir, archiveName string
	weeklyFileWriters               map[time.Time]*os.File
	weeklyTarWriters                map[time.Time]*tar.Writer
	thisSunday, lastSunday          time.Time
	del                             bool
	mutex                           *sync.Mutex
)

func addFile(tw *tar.Writer, thePath string) error {
	mutex.Lock()
	defer mutex.Unlock()
	file, err := os.Open(thePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = path.Base(thePath)
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		// and lock to stop the program closing the tarfile while we're writing to it.
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}

	return nil
}

func getPartNameFromFilepath(thisFile string, sunday time.Time) string {
	name := archiveName
	if archiveName != "" {
		timestamp := utils.TsRegex.FindString(thisFile)
		baseFile := path.Base(thisFile)
		ext := path.Ext(baseFile)
		filename := strings.TrimSuffix(baseFile, ext)
		name = strings.Replace(filename,"_"+timestamp, "", 1)
	}
	datedArchive := sunday.Format(utils.ArchiveForm)
	return fmt.Sprintf(datedArchive, name)+".part"
}

func truncateTimeToSunday(t time.Time) (sunday time.Time) {
	return t.Truncate(time.Hour * 24 * 7)
}

func createNewTar(tarPath string, sunday time.Time){
	var file *os.File
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		file, err = os.Create(tarPath)
		if err != nil {
			errLog.Printf("%s", err)
			panic(err)
		}
	} else {
		file, err = os.OpenFile(tarPath, os.O_RDWR, os.ModePerm)
		if err != nil {
			errLog.Printf("%s", err)
			panic(err)
		}

		if _, err = file.Seek(-2<<9, io.SeekEnd); err != nil {
			errLog.Println(err)
			panic(err)
		}
	}
	weeklyFileWriters[sunday] = file
	weeklyTarWriters[sunday] = tar.NewWriter(file)
	errLog.Printf("[tar] opened %s tar writer", sunday.Format("2006-01-02"))
}

func checkInTar(basePath, tarFileName string, sunday time.Time) bool {
	mutex.Lock()
	defer mutex.Unlock()
	seekpos,_ := weeklyFileWriters[sunday].Seek(0, io.SeekCurrent)

	if _, err := weeklyFileWriters[sunday].Seek(0, io.SeekStart); err != nil {
		errLog.Println(err)
		panic(err)
	}

	defer func() {
		if _, err := weeklyFileWriters[sunday].Seek(seekpos, io.SeekStart); err != nil {
			errLog.Println(err)
			panic(err)
		}
	}()

	reader := tar.NewReader(weeklyFileWriters[sunday])
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			errLog.Println(err)
			panic(err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg, tar.TypeRegA:
			if header.Name == basePath {
				errLog.Println(fmt.Errorf("%s exists in tar file %s", basePath, tarFileName))
				return true
			}
			continue
		default:
			errLog.Println(fmt.Errorf("couldn't determine header Typeflag %s for %s in tar file %s",
				string(header.Typeflag),
				header.Name,
					tarFileName))
		}
	}

	return false
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

	t, err := utils.GetTimeFromFileTimestamp(filePath)
	if err != nil {
		errLog.Println(err)
		return nil
	}
	sunday := truncateTimeToSunday(t)
	if sunday == thisSunday || sunday == lastSunday {
		// dont do anything to this weeks or last weeks files.
		return nil
	}
	basePath := filepath.Base(filePath)
	tarbaseName := getPartNameFromFilepath(filePath, sunday.Add(time.Hour*24*6))
	tarPath := path.Join(outputDir, tarbaseName)


	if _, ok := weeklyTarWriters[sunday]; !ok {
		createNewTar(tarPath, sunday)
	} else {
		inTar := checkInTar(basePath, tarbaseName, sunday)
		if inTar{
			return nil
		}
	}

	if err := addFile(weeklyTarWriters[sunday], filePath); err != nil {
		errLog.Println(err)
		return nil
	}

	if del {
		err := os.Remove(filePath)
		if err != nil {
			errLog.Println(err)
			return nil
		}
	}

	if absPath, err := filepath.Abs(filePath); err == nil {
		utils.EmitPath(absPath)
	} else {
		utils.EmitPath(filePath)
	}
	return nil
}

var usage = func() {
	fmt.Printf("usage of %s:\n", os.Args[0])
	fmt.Println()
	fmt.Println("\tarchive files from directory: ")
	fmt.Printf("\t\t %s -source <source> -output <output>\n", os.Args[0])
	fmt.Println()
	fmt.Println("flags: ")
	fmt.Println()
	fmt.Println("\t-output: set the <destination> directory (default=.)")
	fmt.Println("\t-source: set the <source> directory (optional, default=stdin)")
	fmt.Println("\t-del: delete the source files as they are archived.")
	fmt.Println("\t-name: set the name prefix of the output tarfile <name>~2006-01-02.tar (default=guess)")
	fmt.Println()
	fmt.Println("reads filepaths from stdin")
	fmt.Println("writes paths to resulting files to stdout")
	fmt.Println("will ignore any line from stdin that isnt a filepath (and only a filepath)")
}

func init() {
	errLog = log.New(os.Stderr, "[tsarchive] ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.Usage = usage
	// set flags for flagset
	flag.StringVar(&rootDir, "source", "", "source directory")
	flag.StringVar(&outputDir, "output", "", "output directory")
	flag.StringVar(&archiveName, "name", "", "archive name prefix")
	flag.BoolVar(&del, "del", false, "delete source files")
	// parse the leading argument with normal flag.Parse
	flag.Parse()

	// create dirs
	if rootDir != "" {
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				errLog.Printf("[path] <source> %s does not exist.", rootDir)
				os.Exit(1)
			}
		}
	}
	if outputDir == "" {
		panic(fmt.Errorf("[archive] no output directory specified"))
	}

	if _, err := os.Stat(outputDir); err != nil {
		os.MkdirAll(outputDir, 0750)
	}
	thisSunday = truncateTimeToSunday(time.Now())
	lastSunday = truncateTimeToSunday(time.Now()).Add(-time.Hour * 24 * 7)
}

func main() {
	mutex = &sync.Mutex{}
	weeklyTarWriters = make(map[time.Time]*tar.Writer)
	weeklyFileWriters = make(map[time.Time]*os.File)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		mutex.Lock()
		for sunday, writer := range weeklyTarWriters {
			errLog.Printf("[tar] closing %s tar writer", sunday.Format("2006-01-02"))
			writer.Close()
		}
		for sunday, writer := range weeklyFileWriters {
			errLog.Printf("[tar] closing %s file writer", sunday.Format("2006-01-02"))
			partName := writer.Name()
			writer.Close()
			os.Rename(partName, strings.TrimSuffix(partName, ".part"))
		}
		mutex.Unlock()
		os.Exit(0)
	}()

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
			} else if strings.HasPrefix(text, "#-") {
				// was signalled deletion of previous tmpdir, wait until finished
				defer os.RemoveAll(strings.TrimPrefix(text, "#-"))
			}  else {
				finfo, err := os.Stat(text)
				if err != nil {
					errLog.Printf("[stat] %s", text)
					continue
				}
				visit(text, finfo, nil)
			}
		}
	}
	mutex.Lock()
	for sunday, writer := range weeklyTarWriters {
		errLog.Printf("[tar] closing %s tar writer", sunday.Format("2006-01-02"))
		writer.Close()
	}
	for sunday, writer := range weeklyFileWriters {
		errLog.Printf("[tar] closing %s file writer", sunday.Format("2006-01-02"))
		partName := writer.Name()
		writer.Close()
		os.Rename(partName, strings.TrimSuffix(partName, ".part"))
	}
	mutex.Unlock()
	os.Exit(0)
}
