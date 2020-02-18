# mysqlFindMaster
* A tool to find a master from a specified instance
* Supported Multi Source Replication
* Supported a circular replication of 2 instances
* Not supported a circular replication over 3 instances. Will be error

# Install

```
go get -u github.com/kenken0807/mysqlFindMaster
```

# Use

## Example

```
Master:10.0.0.3 -> Intermediate Slave: 10.0.0.2 -> Slave:10.0.0.1
```

```
#./mysqlFindMaster -h 10.0.0.1 -P 3306 -u UserName -p Password
{"MasterHost":"10.0.0.3","MasterPort":"3306"}
```

## debug mode

```
#./mysqlFindMaster -h 10.0.0.1 -P 3306 -u UserName -p Password -d
Check  Host: 10.0.0.1 Port 3306
Result Host: 10.0.0.1 Port 3306 MasterHost: 10.0.0.2 MasterPort: 3306
Check  Host: 10.0.0.2 Port 3306
Result Host: 10.0.0.2 Port 3306 MasterHost: 10.0.0.3 MasterPort: 3306
Check  Host: 10.0.0.3 Port 3306
IsMaster  Host: 10.0.0.3 Port 3306
{"MasterHost":"10.0.0.3","MasterPort":"3306"}
```


