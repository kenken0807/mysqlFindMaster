package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var Username string
var Password string
var Opt Options
var Debug bool

const VERSION = "1.0.0"
const APPNAME = "mysqlFindMaster"

// Options Info
type Options struct {
	dbUser   string
	dbPasswd string
	host     string
	port     string
	debug    bool
	version  bool
}

func main() {
	Opt = Options{}
	parseOptions(&Opt)
	//version info
	if Opt.version == true {
		fmt.Println(fmt.Sprintf("%s %s", APPNAME, VERSION))
		os.Exit(0)
	}
	// Varidation
	if Opt.host == "" || Opt.port == "" || Opt.dbUser == "" || Opt.dbPasswd == "" {
		fmt.Println("[Error] Specify -h hostname and -P port and -u user -p password")
		os.Exit(1)
	}
	Username = Opt.dbUser
	Password = Opt.dbPasswd
	Debug = Opt.debug

	// set syslog
	logwriter, err := syslog.New(syslog.LOG_NOTICE, APPNAME)
	if err != nil {
		fmt.Println("[Error] at Syslog")
		os.Exit(1)
	}
	log.SetOutput(logwriter)

	findMaster := findMaster{}
	rtn := findMaster.startFindMaster()
	// stdout Master Hostname and Port as JSON
	if rtn == false {
		findMaster.MasterInfos = []MasterInfo{}
		findMaster.masterOutput()
		os.Exit(1)
	}
	findMaster.masterOutput()
	os.Exit(0)
}

func parseOptions(Opt *Options) {
	flag.StringVar(&Opt.dbUser, "u", "", "MySQL Username")
	flag.StringVar(&Opt.dbPasswd, "p", "", "MySQL Password")
	flag.StringVar(&Opt.host, "h", "", "HostName")
	flag.StringVar(&Opt.port, "P", "3306", "MySQL Port")
	flag.BoolVar(&Opt.debug, "d", false, "debug to put log")
	flag.BoolVar(&Opt.version, "v", false, "print version number and exit")
	flag.Parse()

}
