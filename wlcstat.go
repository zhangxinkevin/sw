package sw

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gaochao1/gosnmp"
)

const (
	//AP 
	apNameOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.5"
        apNameOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.5."
	apPowerStatusOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.20"
        apPowerStatusOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.20."
	apAssociatedClientCountOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.54"
        apAssociatedClientCountOidPrefix	 = ".1.3.6.1.4.1.9.9.513.1.1.1.1.54."
	apMemoryCurrentUsageOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.55"
        apMemoryCurrentUsageOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.55."
	apCpuCurrentUsageOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.57"
        apCpuCurrentUsageOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.57."
	apApConnectCountOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.66"
        apApConnectCountOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.66."
	apReassocFailCountOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.68"
        apReassocFailCountOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.68."
	apAssocFailTimesOid	= "1.3.6.1.4.1.9.9.513.1.1.1.1.75"
        apAssocFailTimesOidPrefix	= ".1.3.6.1.4.1.9.9.513.1.1.1.1.75."
	apEthernetIfInputErrorsOid	= "1.3.6.1.4.1.9.9.513.1.2.2.1.17"
        apEthernetIfInputErrorsOidPrefix	 = ".1.3.6.1.4.1.9.9.513.1.2.2.1.17."

)

type WlcStats struct {
        TS                   int64
        ApIndex string
        ApName  string
        ApPowerStatus   int
        ApAssociatedClientCount int
        ApEthernetIfInputErrors int
}

//type WlcStats struct {
//	TS                   int64
//	ApIndex	string
//        ApName	string
//	ApPowerStatus	uint64
//	ApAssociatedClientCount	uint64
//	ApMemoryCurrentUsage	uint64
//	ApCpuCurrentUsage	uint64
//	ApConnectCount	uint64
//	ApReassocFailCount	uint64
//	ApAssocFailTimes	uint64
//	ApEthernetIfInputErrors	uint64
//}

func (this *WlcStats) String() string {
	//return fmt.Sprintf("<IfName:%s, IfIndex:%d, IfHCInOctets:%d, IfHCOutOctets:%d>", this.IfName, this.IfIndex, this.IfHCInOctets, this.IfHCOutOctets)
	return fmt.Sprintf("<ApName:%s, ApIndex:%s>", this.ApName, this.ApIndex)
}

func ListWlcStats(ip, community string, timeout int, ignoreIface []string, retry int, limitConn int, ignorePkt bool, ignoreOperStatus bool, ignoreBroadcastPkt bool, ignoreMulticastPkt bool, ignoreDiscards bool, ignoreErrors bool, ignoreUnknownProtos bool, ignoreOutQLen bool) ([]WlcStats, error) {
	var wlcStatsList []WlcStats
	var limitCh chan bool
	if limitConn > 0 {
		limitCh = make(chan bool, limitConn)
	} else {
		limitCh = make(chan bool, 1)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println(ip+" Recovered in ListWlcStats", r)
		}
	}()
	chApNameList := make(chan []gosnmp.SnmpPDU)
        chApEthernetIfInputErrorsList := make(chan []gosnmp.SnmpPDU)
        chApPowerStatusList := make(chan []gosnmp.SnmpPDU)

        limitCh <- true
	go ListApName(ip, community, timeout, chApNameList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)
        limitCh <- true
        go ListApEthernetIfInputErrors(ip, community, timeout, chApEthernetIfInputErrorsList, retry, limitCh)
        time.Sleep(5 * time.Millisecond)
        limitCh <- true
        go ListApPowerStatus(ip, community, timeout, chApPowerStatusList, retry, limitCh)
        time.Sleep(5 * time.Millisecond)

	apNameList := <-chApNameList
	apEthernetIfInputErrorsList := <-chApEthernetIfInputErrorsList
	apPowerStatusList := <-chApPowerStatusList

	if len(apNameList) > 0 && len(apPowerStatusList) > 0 {
		now := time.Now().Unix()

		for _, apNamePDU := range apNameList {

			apName := apNamePDU.Value.(string)

			check := true
			if check {
				var wlcStats WlcStats
				apIndexStr := strings.Replace(apNamePDU.Name, apNameOidPrefix, "", 1)

				wlcStats.ApIndex = apIndexStr

				for ti, apPowerStatusPDU := range apPowerStatusList {
					if strings.Replace(apPowerStatusPDU.Name, apPowerStatusOidPrefix, "", 1) == apIndexStr {
						wlcStats.ApPowerStatus = apPowerStatusList[ti].Value.(int)
						break
					}
				}	
                                for ti, apEthernetIfInputErrorsPDU := range apEthernetIfInputErrorsList {
                                        if strings.Replace(apEthernetIfInputErrorsPDU.Name, apEthernetIfInputErrorsOidPrefix, "", 1) == apIndexStr {
                                                wlcStats.ApEthernetIfInputErrors = apEthernetIfInputErrorsList[ti].Value.(int)
                                                break
                                        }
                                }
				wlcStats.TS = now
				wlcStats.ApName = apName
				wlcStatsList = append(wlcStatsList, wlcStats)
			}
		}
	}
        return wlcStatsList, nil
}

func ListApName(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
        RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, apNameOid)
}

func ListApPowerStatus(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
        RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, apPowerStatusOid)
}
func ListApEthernetIfInputErrors(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
        RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, apEthernetIfInputErrorsOid)
}



//func RunSnmpRetry(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool, oid string) {
//
//	var snmpPDUs []gosnmp.SnmpPDU
//	var err error
//	snmpPDUs, err = RunSnmpwalk(ip, community, oid, retry, timeout)
//	if err != nil {
//		log.Println(ip, oid, err)
//		close(ch)
//		<-limitCh
//		return
//	}
//	<-limitCh
//	ch <- snmpPDUs
//
//	return
//}
