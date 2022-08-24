package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/google/fscrypt/filesystem"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	mountpoint = kingpin.Flag("mount", "The mount to find the largest file usages. Can be a subath of mount").Required().String()
	limit      = kingpin.Flag("limit", "The maximum number of files return to the display").Default("10").Short('l').Int()
)
var device string

type fileDisplays struct {
	sync.RWMutex
	Files []fileDisplay
}

var files fileDisplays

type fileDisplay struct {
	Size int64
	Path string
}
type bySize []fileDisplay

func (a bySize) Len() int           { return len(a) }
func (a bySize) Less(i, j int) bool { return a[i].Size < a[j].Size }
func (a bySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {
	log.SetOutput(io.Discard)
	kingpin.Version("0.0.1")
	kingpin.Parse()
	//Define limit after parsing
	logrus.SetLevel(logrus.FatalLevel)
	if (*mountpoint)[len(*mountpoint)-1:] != "/" {
		*mountpoint = *mountpoint + "/"
	}
	fmt.Println("Finding the top", *limit, "largest files on filesystem", *mountpoint, "\n================================================")
	mount, err := filesystem.FindMount(*mountpoint)
	if err != nil {
		logrus.Fatal(err)
	}
	device = mount.Device

	entries, err := os.ReadDir(*mountpoint)
	if err != nil {
		logrus.Fatal(err)
	}
	var wg sync.WaitGroup
	getFiles(*mountpoint, entries, &wg)
	wg.Wait()
	sort.Sort(bySize(files.Files))
	var shortFiles []fileDisplay
	if len(files.Files) > *limit {
		shortFiles = files.Files[len(files.Files)-*limit:]
	} else {
		shortFiles = files.Files
	}

	for _, file := range shortFiles {
		fmt.Println(file.Path, file.DisplaySizeIEC())
	}

}

func getFiles(start string, entries []fs.DirEntry, wg *sync.WaitGroup) {
	for _, entry := range entries {
		wg.Add(1)
		go handleEntry(start, entry, wg)
	}

}

func handleEntry(start string, entry fs.DirEntry, wg *sync.WaitGroup) {
	defer wg.Done()
	var file fileDisplay
	mount, err := filesystem.FindMount(start + entry.Name())
	if err != nil {
		logrus.Errorln(err, start+entry.Name())
		return
	}
	if mount.Device == device {
		if entry.Type().IsRegular() {
			fileInfo, err := os.Stat(start + entry.Name())
			if err != nil {
				logrus.Errorln(err, start+entry.Name())
				return
			}
			file.Path = start + entry.Name()
			file.Size = fileInfo.Size()
			files.Append(file)
		} else if entry.IsDir() {
			entries, err := os.ReadDir(start + entry.Name())
			if err != nil {
				logrus.Errorln(err, start+entry.Name())
				return
			}
			logrus.Info("Searching ", start+entry.Name())
			getFiles(start+entry.Name()+"/", entries, wg)
		}
	}

}

func (f *fileDisplay) DisplaySizeIEC() string {
	const unit = 1024
	b := f.Size
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f%ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func (fd *fileDisplays) Append(item fileDisplay) {
	fd.Lock()
	defer fd.Unlock()

	fd.Files = append(fd.Files, item)
}
