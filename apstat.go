package sw

import (
	"log"
	"time"

	"github.com/gaochao1/gosnmp"
)

func ApJoinStatus(ip, community string, timeout, retry int) (int, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(ip+" Recovered in CPUtilization", r)
		}
	}()
	vendor, err := SysVendor(ip, community, retry, timeout)
	if err != nil {
		return 0, err
	}
	method := "get"
	var oid string

	switch vendor {
	case "Cisco_WCL":
		oid = "1.3.6.1.4.1.9.9.618.1.8.4.0"
	}

	var snmpPDUs []gosnmp.SnmpPDU
	for i := 0; i < retry; i++ {
		snmpPDUs, err = RunSnmp(ip, community, oid, method, timeout)
		if len(snmpPDUs) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err == nil {
		for _, pdu := range snmpPDUs {
			return pdu.Value.(int), err
		}
	}

	return 0, err
}
