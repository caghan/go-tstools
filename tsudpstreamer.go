package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/aristanetworks/goarista/atime" //for getting monotonic clock time
)

const (
	maxDatagramSize  = 1316
	TSPacketSize     = 188
	fileChunkSize    = 500
	streamBufferSize = 65535
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Command line udp streamer with arguments in order:
// # tsudpstreamer 224.0.1.2:10001 eth0 test.ts 2496916
func main() {
	if len(os.Args) == 5 {
		srvAddr, ifName, fName := os.Args[1], os.Args[2], os.Args[3]
		bitrate, err := strconv.ParseUint(os.Args[4], 10, 64)
		check(err)
		startUDPStream(srvAddr, fName, ifName, bitrate)
	} else {
		fmt.Fprintf(os.Stderr, "usage: %s [multicastAddress:Port] [ifname] [inputTSfile] [TSbitrate] \n", os.Args[0])
	}
}

//returns UDP Connection for the given network device name (network interface)
func getUDPConnection(ifName string, bindAddr string) (*net.UDPConn, error) {
	netDev, err := net.InterfaceByName(ifName)
	check(err)
	addrs, err := netDev.Addrs()
	check(err)
	lAddr := &net.UDPAddr{
		IP: addrs[0].(*net.IPNet).IP,
	}
	addr, err := net.ResolveUDPAddr("udp", bindAddr)
	check(err)
	return net.DialUDP("udp", lAddr, addr)
}

func startUDPStream(bindAddr string, fName string, ifName string, bitrate uint64) {

	conn, err := getUDPConnection(ifName, bindAddr)
	check(err)
	packetSize := 7 * TSPacketSize
	file, err := os.Open(fName)
	check(err)
	defer file.Close()

	buf := make([]byte, packetSize)

	completed := 0
	packetTime, timeStart, timeStop, realTime := uint64(0), uint64(0), uint64(0), uint64(0)

	nanoSleepPacket := syscall.Timespec{}
	nanoSleepPacket.Nsec = 665778 // 1 packet at 100mbps

	timeStart = atime.NanoTime()

	//conn.SetWriteBuffer(229376)
	conn.SetWriteBuffer(streamBufferSize)
	for completed != 1 {
		timeStop = atime.NanoTime()
		realTime = (timeStop - timeStart)
		if realTime*bitrate/1000 > packetTime*1000000 && completed != 1 {
			tmp, err := file.Read(buf)
			if err != nil {
				if err == io.EOF {
					completed = 1
				} else {
					panic(err)
				}
			}
			if tmp <= 0 {
				completed = 1
			} else {
				//log.Println("bytes: ", tmp, ", string: ", hex.Dump(buf[:tmp]))
				conn.Write(buf[:tmp])
				packetTime += uint64(packetSize * 8)
			}
		} else {
			syscall.Nanosleep(&nanoSleepPacket, nil) //works only on Linux
			//time.Sleep(1 * time.Second) //substitute for ~MS & MacOSX
		}
	}
}
