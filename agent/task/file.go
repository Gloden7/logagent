package task

import (
	"logagent/tail"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func getLatestFile(dir string, r *regexp.Regexp) string {
	latestFilePath := ""
	latestFileModTime := time.Unix(0, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && r.MatchString(info.Name()) && info.ModTime().After(latestFileModTime) {
			latestFilePath = path
		}
		return nil
	})
	return latestFilePath
}

func getTailf(p string) (*tail.Tail, error) {
	return tail.TailFile(p, tail.Config{
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: true,
	})
}
