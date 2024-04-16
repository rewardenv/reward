package globals

var (
	APPNAME string
	VERSION string
)

func InitGlobals(name string, version string) {
	APPNAME = name
	VERSION = version
}
