package sw

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gaochao1/gosnmp"
)

const (
	acifNameOid             = "1.3.6.1.4.1.35047.2.1.2.1.2"
	acifNameOidPrefix       = ".1.3.6.1.4.1.35047.2.1.2.1.2."
	acifHCInOid             = "1.3.6.1.4.1.35047.2.1.2.1.7"
	acifHCInOidPrefix       = ".1.3.6.1.4.1.35047.2.1.2.1.7."
	acifHCOutOid            = "1.3.6.1.4.1.35047.2.1.2.1.8"
	acifHCOutOidPrefix      = ".1.3.6.1.4.1.35047.2.1.2.1.8."
	acifOperStatusOid       = "1.3.6.1.4.1.35047.2.1.2.1.4"
	acifOperStatusOidPrefix = ".1.3.6.1.4.1.35047.2.1.2.1.4."
)

type AcIfStats struct {
	AcIfName        string
	AcIfHCInOctets  int
	AcIfHCOutOctets int
	AcIfOperStatus  string
	TS              int64
}

func (this *AcIfStats) String() string {
	return fmt.Sprintf("<IfName:%s, IfHCInOctets:%d, IfHCOutOctets:%d, IfOperStatus:%s>, ", this.AcIfName, this.AcIfHCInOctets, this.AcIfHCOutOctets, this.AcIfOperStatus)
}

func ListAcIfStats(ip, community string, timeout int, ignoreIface []string, retry int, limitConn int, ignorePkt bool, ignoreOperStatus bool, ignoreBroadcastPkt bool, ignoreMulticastPkt bool, ignoreDiscards bool, ignoreErrors bool, ignoreUnknownProtos bool, ignoreOutQLen bool) ([]AcIfStats, error) {
	var acIfStatsList []AcIfStats
	var limitCh chan bool
	if limitConn > 0 {
		limitCh = make(chan bool, limitConn)
	} else {
		limitCh = make(chan bool, 1)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println(ip+" Recovered in ListIfStats", r)
		}
	}()

	chAcIfInList := make(chan []gosnmp.SnmpPDU)
	chAcIfOutList := make(chan []gosnmp.SnmpPDU)
	chAcIfNameList := make(chan []gosnmp.SnmpPDU)

	limitCh <- true
	go ListAcIfHCInOctets(ip, community, timeout, chAcIfInList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)
	limitCh <- true
	go ListAcIfHCOutOctets(ip, community, timeout, chAcIfOutList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)
	limitCh <- true
	go ListAcIfName(ip, community, timeout, chAcIfNameList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)

	// OperStatus
	var acIfStatusList []gosnmp.SnmpPDU
	chAcIfStatusList := make(chan []gosnmp.SnmpPDU)
	if ignoreOperStatus == false {
		limitCh <- true
		go ListAcIfOperStatus(ip, community, timeout, chAcIfStatusList, retry, limitCh)
		time.Sleep(5 * time.Millisecond)
	}
	acIfInList := <-chAcIfInList
	acIfOutList := <-chAcIfOutList
	acIfNameList := <-chAcIfNameList
	if ignoreOperStatus == false {
		acIfStatusList = <-chAcIfStatusList
	}
	if len(acIfNameList) > 0 && len(acIfInList) > 0 && len(acIfOutList) > 0 {
		now := time.Now().Unix()

		for _, ifNamePDU := range acIfNameList {

			acIfName := ifNamePDU.Value.(string)

			check := true
			if len(ignoreIface) > 0 {
				for _, ignore := range ignoreIface {
					if strings.Contains(acIfName, ignore) {
						check = false
						break
					}
				}
			}

			if check {
				var acIfStats AcIfStats
				ifIndexStr := strings.Replace(ifNamePDU.Name, acifNameOidPrefix, "", 1)

				for ti, acIfHCInOctetsPDU := range acIfInList {

					if strings.Replace(acIfHCInOctetsPDU.Name, acifHCInOidPrefix, "", 1) == ifIndexStr {
						acIfStats.AcIfHCInOctets = acIfInList[ti].Value.(int)
						break
					}
				}

				for ti, acifHCOutOidPDU := range acIfOutList {
					if strings.Replace(acifHCOutOidPDU.Name, acifHCOutOidPrefix, "", 1) == ifIndexStr {
						acIfStats.AcIfHCOutOctets = acIfOutList[ti].Value.(int)
						break
					}
				}

				if ignoreOperStatus == false {
					for ti, acifOperStatusPDU := range acIfStatusList {
						if strings.Replace(acifOperStatusPDU.Name, acifOperStatusOidPrefix, "", 1) == ifIndexStr {
							acIfStats.AcIfOperStatus = acIfStatusList[ti].Value.(string)
							break
						}
					}
				}

				acIfStats.TS = now
				acIfStats.AcIfName = acIfName
				acIfStatsList = append(acIfStatsList, acIfStats)
			}
		}
	}

	return acIfStatsList, nil
}

func ListAcIfOperStatus(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, acifOperStatusOid)
}

func ListAcIfName(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, acifNameOid)
}

func ListAcIfHCInOctets(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, acifHCInOid)
}

func ListAcIfHCOutOctets(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, acifHCOutOid)
}
