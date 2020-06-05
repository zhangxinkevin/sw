package sw

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gaochao1/gosnmp"
)

const (
	//AC && AD
	diskNameOid              = "1.3.6.1.4.1.35047.1.5.1.2"
	diskNameOidPrefix        = ".1.3.6.1.4.1.35047.1.5.1.2."
	diskAvailOid             = "1.3.6.1.4.1.35047.1.5.1.5"
	diskAvailOidPrefix       = ".1.3.6.1.4.1.35047.1.5.1.5."
	diskUsedPercentOid       = "1.3.6.1.4.1.35047.1.5.1.6"
	diskUsedPercentOidPrefix = ".1.3.6.1.4.1.35047.1.5.1.6."
)

type DiskStats struct {
	DiskName        string
	DiskAvail       string
	DiskUsedPercent string
	TS              int64
}

func (this *DiskStats) String() string {
	//return fmt.Sprintf("<IfName:%s, IfIndex:%d, IfHCInOctets:%d, IfHCOutOctets:%d>", this.IfName, this.IfIndex, this.IfHCInOctets, this.IfHCOutOctets)
	return fmt.Sprintf("<AcName:%s, AcIndex:%s>", this.DiskName, this.DiskAvail)
}

func ListDiskStats(ip, community string, timeout int, ignoreIface []string, retry int, limitConn int, ignorePkt bool, ignoreOperStatus bool, ignoreBroadcastPkt bool, ignoreMulticastPkt bool, ignoreDiscards bool, ignoreErrors bool, ignoreUnknownProtos bool, ignoreOutQLen bool) ([]DiskStats, error) {
	var diskStatsList []DiskStats
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
	chDiskNameList := make(chan []gosnmp.SnmpPDU)
	chDiskAvailList := make(chan []gosnmp.SnmpPDU)
	chDiskUsedPercentList := make(chan []gosnmp.SnmpPDU)

	limitCh <- true
	go ListDiskName(ip, community, timeout, chDiskNameList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)
	limitCh <- true
	go ListDiskAvail(ip, community, timeout, chDiskAvailList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)
	limitCh <- true
	go ListdiskUsedPercent(ip, community, timeout, chDiskUsedPercentList, retry, limitCh)
	time.Sleep(5 * time.Millisecond)

	diskNameList := <-chDiskNameList
	diskAvailList := <-chDiskAvailList
	diskUsedPercentList := <-chDiskUsedPercentList

	if len(diskNameList) > 0 {
		now := time.Now().Unix()

		for _, diskNamePDU := range diskNameList {
			diskName := diskNamePDU.Value.(string)

			check := true
			// if len(ignoreIface) > 0 {
			// 	for _, ignore := range ignoreIface {
			// 		if strings.Contains(diskName, ignore) {
			// 			check = false
			// 			break
			// 		}
			// 	}
			// }
			if check {
				var diskStats DiskStats
				diskIndexStr := strings.Replace(diskNamePDU.Name, diskNameOidPrefix, "", 1)

				for ti, diskAvailPDU := range diskAvailList {
					if strings.Replace(diskAvailPDU.Name, diskAvailOidPrefix, "", 1) == diskIndexStr {
						diskStats.DiskAvail = diskAvailList[ti].Value.(string)
						break
					}
				}

				for ti, diskUsedPercentPDU := range diskUsedPercentList {
					if strings.Replace(diskUsedPercentPDU.Name, diskUsedPercentOidPrefix, "", 1) == diskIndexStr {
						diskStats.DiskUsedPercent = diskUsedPercentList[ti].Value.(string)
						break
					}
				}

				diskStats.TS = now
				diskStats.DiskName = diskName
				diskStatsList = append(diskStatsList, diskStats)
			}
		}
	}
	return diskStatsList, nil
}

func ListDiskName(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, diskNameOid)
}

func ListDiskAvail(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, diskAvailOid)
}

func ListdiskUsedPercent(ip, community string, timeout int, ch chan []gosnmp.SnmpPDU, retry int, limitCh chan bool) {
	RunSnmpRetry(ip, community, timeout, ch, retry, limitCh, diskUsedPercentOid)
}
