package common

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/zput/zxcTool/ztLog/zt_formatter"
)

func SetupLog(fields_order ...string) {
	log.SetReportCaller(true)
	log.SetFormatter(&zt_formatter.ZtFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			var filename string = path.Base(f.File)
			dirs := strings.Split(path.Dir(f.File), string(os.PathSeparator))
			if len(dirs) > 0 {
				base_dir := dirs[len(dirs)-1]
				filename = strings.Join([]string{base_dir, filename}, string(os.PathSeparator))
			}
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
		Formatter: nested.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			ShowFullLevel:   true,
			FieldsOrder:     fields_order,
		},
	})
}

func SetLevel(level string) error {
	l, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(l)
	return nil
}
