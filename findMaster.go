package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type slaveStatusInfo struct {
	masterHost string
	masterPort string
	ioThread   string
	sqlThread  string
}

type dbInfo struct {
	hostName   string
	port       string
	db         *sql.DB
	readOnly   int
	isMaster   bool
	slaveCount int
	slaveInfos []slaveStatusInfo
}

type dbLists struct {
	checkedHostname map[string]bool
	dbInfos         []dbInfo
}

type findMaster struct {
	multiSlaveCnt int
	MasterInfos   []MasterInfo
}

// MasterInfo ..
type MasterInfo struct {
	MasterHost string
	MasterPort string
}

func (findMaster *findMaster) startFindMaster() bool {
	// Initial dbinfo
	hostName := Opt.host
	port := Opt.port
	findMaster.MasterInfos = []MasterInfo{}

	// ipv4 or not
	var ok bool
	hostName, ok = validateHostname(hostName)
	if !ok {
		masterErrInfo(fmt.Sprintf("[ERROR]%s is not IP Address or DNS Record", hostName))
		return false
	}
	if ok := findMaster.followMaster(hostName, port); !ok {
		return false
	}
	return true

}

func (findMaster *findMaster) followMaster(hostName string, port string) bool {
	dbLists := dbLists{}
	dbLists.dbInfos = []dbInfo{}
	dbLists.checkedHostname = make(map[string]bool, 0)
	dbInfos := dbLists.dbInfos
	dbInfos = append(dbInfos, dbInfo{hostName: hostName, port: port})

	// Start Loop
	i := 0
	for {
		dbInfos[i].slaveInfos = []slaveStatusInfo{}
		slaveInfos := dbLists.findMaster(&dbInfos[i])
		if dbInfos[i].isMaster == true {
			// Master Found
			findMaster.setMasterHostPort(dbInfos[i].hostName, dbInfos[i].port)
			return true
		}
		if slaveInfos == nil {
			// Something Error
			return false
		}
		for idx, slaveInfo := range slaveInfos {
			if idx == 0 {
				dbInfos = append(dbInfos, dbInfo{hostName: slaveInfo.masterHost, port: slaveInfo.masterPort})
			} else {
				if ok := findMaster.followMaster(slaveInfo.masterHost, slaveInfo.masterPort); !ok {
					return false
				}
			}
		}
		if len(dbInfos) > 2 {
			// Check for circular replication
			previous := &dbInfos[len(dbInfos)-3]
			latest := &dbInfos[len(dbInfos)-2]
			if err := findMaster.electMaster(previous, latest); err != nil {
				// Error Find Master for circular replication
				return false
			}
			// Find Master for circular replication
			if previous.isMaster == true {
				findMaster.setMasterHostPort(previous.hostName, previous.port)
				return true
			}
			if latest.isMaster == true {
				findMaster.setMasterHostPort(latest.hostName, latest.port)
				return true
			}
		}
		i++
		// Prevent infinity loops for safety
		if i > 100 {
			masterErrInfo(fmt.Sprintf("[ERROR] Prevent Infinity Loops. Checked instance count is[%d]", i))
			return false
		}
	}
}

func (findMaster *findMaster) setMasterHostPort(hostName string, port string) {
	m.Lock()
	defer m.Unlock()
	cnt := 0
	for _, masterinfo := range findMaster.MasterInfos {
		if masterinfo.MasterHost == hostName && masterinfo.MasterPort == port {
			cnt++
		}
	}
	if cnt == 0 {
		findMaster.MasterInfos = append(findMaster.MasterInfos, MasterInfo{MasterHost: hostName, MasterPort: port})
		masterErrInfo(fmt.Sprintf("IsMaster  Host: %s Port %s", hostName, port))

	} else {
		masterErrInfo(fmt.Sprintf("Choose Master is already added, so ignore Host: %s Port %s", hostName, port))
	}
}

func (findMaster *findMaster) masterOutput() {
	bytes, _ := json.Marshal(findMaster.MasterInfos)
	fmt.Println(string(bytes))
}

func (findMaster *findMaster) electMaster(previous *dbInfo, latest *dbInfo) error {
	// for multi Replication
	// previous == latest masterHost and previous masterHost == latest is multi replication
	if previous.hostName == latest.slaveInfos[0].masterHost && previous.port == latest.slaveInfos[0].masterPort &&
		latest.hostName == previous.slaveInfos[0].masterHost && latest.port == previous.slaveInfos[0].masterPort {
		masterErrInfo(fmt.Sprintf("Elect Master - Host: %s Port %s Read_Only=%d SlaveCount: %d,Host: %s Port %s Read_Only=%d SlaveCount: %d",
			previous.hostName, previous.port, previous.readOnly, previous.slaveCount, latest.hostName, latest.port, latest.readOnly, latest.slaveCount))
		if previous.readOnly == 0 && latest.readOnly == 1 {
			previous.isMaster = true
			return nil
		}
		if previous.readOnly == 1 && latest.readOnly == 0 {
			latest.isMaster = true
			return nil
		}
		// SHOW PROCESSLIST CHECK
		if previous.readOnly == 0 && latest.readOnly == 0 {
			if previous.slaveCount > latest.slaveCount {
				previous.isMaster = true
				return nil
			}
			if previous.slaveCount < latest.slaveCount {
				latest.isMaster = true
				return nil
			}
			if previous.slaveCount == latest.slaveCount {
				previous.isMaster = true
				return nil
			}

		}
		if previous.readOnly == 1 && latest.readOnly == 1 {
			buf := fmt.Sprintf("[ERROR] Host1: %s:%s Host2: %s:%s Message: %s", previous.hostName, previous.port, latest.hostName, latest.port, "Both ReadOnly=ON, so nothing to do")
			masterErrInfo(buf)
			return fmt.Errorf(buf)
		}
	}
	// maybe cascade replication, so check again
	return nil
}
