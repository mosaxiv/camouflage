package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port              string
	SharedKey         string
	HeaderVia         string
	LengthLimit       int64
	LogingEnabled     string
	MaxRedirects      int
	TimingAllowOrigin string
	HostName          string
	Timeout           time.Duration
	DisableKeepAlive  bool
	MimeTypes         []string
}

var defaultMimeTypes = []string{
	"image/bmp",
	"image/cgm",
	"image/g3fax",
	"image/gif",
	"image/ief",
	"image/jp2",
	"image/jpeg",
	"image/jpg",
	"image/pict",
	"image/png",
	"image/prs.btif",
	"image/svg+xml",
	"image/tiff",
	"image/vnd.adobe.photoshop",
	"image/vnd.djvu",
	"image/vnd.dwg",
	"image/vnd.dxf",
	"image/vnd.fastbidsheet",
	"image/vnd.fpx",
	"image/vnd.fst",
	"image/vnd.fujixerox.edmics-mmr",
	"image/vnd.fujixerox.edmics-rlc",
	"image/vnd.microsoft.icon",
	"image/vnd.ms-modi",
	"image/vnd.net-fpx",
	"image/vnd.wap.wbmp",
	"image/vnd.xiff",
	"image/webp",
	"image/x-cmu-raster",
	"image/x-cmx",
	"image/x-icon",
	"image/x-macpaint",
	"image/x-pcx",
	"image/x-pict",
	"image/x-portable-anymap",
	"image/x-portable-bitmap",
	"image/x-portable-graymap",
	"image/x-portable-pixmap",
	"image/x-quicktime",
	"image/x-rgb",
	"image/x-xbitmap",
	"image/x-xpixmap",
	"image/x-xwindowdump",
}

func NewConfig() Config {
	return Config{
		Port:              getEnvStr("PORT", "8081"),
		SharedKey:         getEnvStr("CAMO_KEY", "0x24FEEDFACEDEADBEEFCAFE"),
		HeaderVia:         getEnvStr("CAMO_HEADER_VIA", "Camo Asset Proxy"),
		LengthLimit:       getEnvInt64("CAMO_LENGTH_LIMIT", 5242880),
		LogingEnabled:     getEnvStr("CAMO_LOGGING_ENABLED", "disabled"),
		MaxRedirects:      getEnvInt("CAMO_MAX_REDIRECTS", 4),
		TimingAllowOrigin: getEnvStr("CAMO_TIMING_ALLOW_ORIGIN", ""),
		HostName:          getEnvStr("CAMO_HOSTNAME", "unknown"),
		Timeout:           time.Duration(getEnvInt("CAMO_SOCKET_TIMEOUT", 10)) * time.Second,
		DisableKeepAlive:  !getEnvBool("CAMO_KEEP_ALIVE", false),
		MimeTypes:         getMimeTypes(),
	}
}

func getMimeTypes() []string {
	var mimeTypes []string
	if _, err := os.Stat("mime-types.json"); err == nil {
		bytes, err := ioutil.ReadFile("mime-types.json")
		if err != nil {
			log.Fatalln(err.Error())
		}
		if err := json.Unmarshal(bytes, &mimeTypes); err != nil {
			log.Fatalln(err.Error())
		}
		return mimeTypes
	}

	return defaultMimeTypes
}

func getEnvStr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("%s : %s", key, err.Error())
	}

	return i
}

func getEnvInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Fatalf("%s : %s", key, err.Error())
	}

	return i
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	switch strings.ToLower(v) {
	case "true", "1":
		return true
	case "false", "0":
		return false
	default:
		return false
	}
}
