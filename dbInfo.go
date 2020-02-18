package main

import (
	"database/sql"
	"fmt"
)

func (dbLists *dbLists) findMaster(dbInfo *dbInfo) []slaveStatusInfo {

	if Debug {
		fmt.Println(fmt.Sprintf("Check  Host: %s Port %s", dbInfo.hostName, dbInfo.port))
	}
	// An error occurs when over 3 instances circular Replication
	if ok := dbLists.setCheckedHostPort(dbInfo.hostName, dbInfo.port); !ok {
		return nil
	}
	// Find SlaveInfo and InstanceInfo
	if ok := dbInfo.checkInstanceInfo(); !ok {
		return nil
	}
	// This is a Master. No show slave status Info.
	if dbInfo.isMaster == true {
		return nil
	}
	if Debug {
		if len(dbInfo.slaveInfos) > 1 {
			fmt.Println(fmt.Sprintf("MultiSourceReplication  Host: %s Port %s", dbInfo.hostName, dbInfo.port))
		}
		for _, slaveInfo := range dbInfo.slaveInfos {
			fmt.Println(fmt.Sprintf("Result Host: %s Port %s MasterHost: %s MasterPort: %s", dbInfo.hostName, dbInfo.port, slaveInfo.masterHost, slaveInfo.masterPort))
		}
	}
	return dbInfo.slaveInfos
}

func (dbLists *dbLists) setCheckedHostPort(host string, port string) bool {
	hostPort := fmt.Sprintf("%s:%s", host, port)
	if dbLists.checkedHostname[hostPort] == false {
		dbLists.checkedHostname[hostPort] = true
		return true
	}
	masterErrInfo(fmt.Sprintf("[ERROR] Host: %s:%s is already checked. It might be multi repication over 3 instances", host, port))
	return false
}

func (dbInfo *dbInfo) checkInstanceInfo() bool {
	// target DB connect
	if err := dbInfo.connection(); err != nil {
		dbInfo.masterErrInfo("connection()", err)
		return false
	}
	// @@READ_ONLY or not
	if err := dbInfo.getReadOnly(); err != nil {
		dbInfo.masterErrInfo("getReadOnly()", err)
		return false
	}
	// How Many Slave have
	if err := dbInfo.getSlaveCount(); err != nil {
		dbInfo.masterErrInfo("getSlaveCount()", err)
		return false
	}
	// Get SHOW SLAVE STATUS INFO
	if err := dbInfo.getSlaveStatus(); err != nil {
		dbInfo.masterErrInfo("getSlaveStatus()", err)
		return false
	}
	// No show slave info means Master
	if len(dbInfo.slaveInfos) == 0 {
		// Master have to be READ_ONLY=OFF
		if ok, comment := dbInfo.masterStatusCheck(); !ok {
			dbInfo.masterErrInfo(comment, nil)
			return false
		}
		dbInfo.isMaster = true
		return true

	}
	// To be Error if IO/SQL Thread not Yes
	for _, slaveInfo := range dbInfo.slaveInfos {
		if ok, reason := slaveInfo.statusCheck(); !ok {
			dbInfo.masterErrInfo(reason, nil)
			return false
		}
	}
	return true
}

func (dbInfo *dbInfo) getSlaveStatus() error {
	var rows *sql.Rows
	var err error
	rows, err = dbInfo.db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	for rows.Next() {
		scanArgs := make([]interface{}, len(cols))
		for i := range scanArgs {
			scanArgs[i] = &sql.RawBytes{}
		}
		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}
		masterHost := columnValue(scanArgs, cols, "Master_Host")
		// ipv4 or not
		var ok bool
		masterHost, ok = validateHostname(masterHost)
		if !ok {
			dbInfo.masterErrInfo("dnsNameResolve()", nil)
			return fmt.Errorf("Error: %s", "dnsNameResolve()")
		}
		dbInfo.slaveInfos = append(dbInfo.slaveInfos, slaveStatusInfo{masterHost: masterHost,
			masterPort: columnValue(scanArgs, cols, "Master_Port"),
			ioThread:   columnValue(scanArgs, cols, "Slave_IO_Running"),
			sqlThread:  columnValue(scanArgs, cols, "Slave_SQL_Running")})
	}
	return nil
}

func (dbInfo *dbInfo) masterErrInfo(buf string, err error) {
	masterErrInfo(fmt.Sprintf("[ERROR] ConnectHost: %s, ConnectPort: %s Message: %s Error: %v", dbInfo.hostName, dbInfo.port, buf, err))
}

func (dbInfo *dbInfo) masterStatusCheck() (bool, string) {
	if dbInfo.readOnly == 1 {
		return false, fmt.Sprint("Read_Only is ON")
	}
	return true, ""
}

func (dbInfo *dbInfo) getReadOnly() error {
	if err := dbInfo.db.QueryRow("SELECT @@READ_ONLY").Scan(&dbInfo.readOnly); err != nil {
		return err
	}
	return nil
}

func (dbInfo *dbInfo) getSlaveCount() error {
	if err := dbInfo.db.QueryRow("SELECT COUNT(*) FROM information_schema.processlist WHERE command IN ('Binlog Dump', 'Binlog Dump GTID')").Scan(&dbInfo.slaveCount); err != nil {
		return err
	}
	return nil
}

func (dbInfo *dbInfo) connection() error {
	var err error
	dbInfo.db, err = dbconn(dbInfo.hostName, dbInfo.port, Username, Password, "mysql")
	if err != nil {

		return err
	}
	return nil
}

func dbconn(host string, port string, dbuser string, dbpasswd string, dbDB string) (*sql.DB, error) {
	return sql.Open("mysql", dbuser+":"+dbpasswd+"@tcp("+host+":"+port+")/"+dbDB+"?readTimeout=2s&timeout=2s")
}
