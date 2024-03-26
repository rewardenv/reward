package globals

var APPNAME string
var VERSION string

func InitGlobals(name string, version string) {
	APPNAME = name
	VERSION = version
}
