package sw

import (
	"errors"
	"log"
	"time"

	"github.com/gaochao1/gosnmp"
)

func CpuMgmtUtilization(ip, community string, timeout, retry int) (int, error) {
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
	case "PA_800", "PA":
		oid = "1.3.6.1.2.1.25.3.3.1.2.1"
	default:
		err = errors.New(ip + " Switch Vendor is not defined")
		return 0, err
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
