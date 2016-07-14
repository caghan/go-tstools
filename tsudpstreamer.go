package main

import (
	"io"
	"log"
	"net"
	"os"
	"github.com/aristanetworks/goarista/atime" //for getting monotonic clock time
)

const (
	MULTICAST_ADDR = "127.0.0.1:12345"
	TS_FILE_NAME = "test.ts"
	MPEG_TS_BITRATE = 9358043
	TS_PACKET_SIZE = 188
	TS_BUFFER_SIZE = TS_PACKET_SIZE * 8 // 1316 for UDP streaming
	//ONE_TS_PACKET_NS_FOR_100MPS = 665778 // 1 packet at 100mbps
	LOOP_TS = true
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.Println("Streaming of ", TS_FILE_NAME, "started!")
	udpStreamTo(MULTICAST_ADDR)
	log.Println("Streaming of ", TS_FILE_NAME, "ended!")
}

func udpStreamTo(multicastAddr string) {
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	checkError(err)
	udpConn, err := net.DialUDP("udp", nil, addr)
	checkError(err)
	tsFile, err := os.Open(TS_FILE_NAME)
	checkError(err)
	tsBuffer := make([]byte, TS_BUFFER_SIZE)

	completed := 0
	packetTime := uint64(0)
	timeStart := uint64(0)
	timeStop := uint64(0)
	realTime := uint64(0)

	timeStart = atime.NanoTime()

	for completed != 1 {
		timeStop = atime.NanoTime()
		realTime = (timeStop - timeStart) / 1000
		if realTime* MPEG_TS_BITRATE > packetTime*1000000 && completed != 1 {
			chunk, err := tsFile.Read(tsBuffer)
			if err != nil {
				if err == io.EOF {
					completed = 1
					if LOOP_TS {
						udpStreamTo(multicastAddr)
					}
				} else {
					panic(err)
				}
			}
			if chunk < 0 {
				log.Println("TS FILE SENT!")
				completed = 1
			} else if chunk == 0 {
				log.Println("TS FILE READ ERROR!")
				completed = 1
			} else {
				udpConn.Write(tsBuffer[:chunk])
				packetTime += uint64(TS_BUFFER_SIZE * 8)
			}
		} else {
			//syscall.Nanosleep(&ONE_TS_PACKET_NS_FOR_100MPS, nil) // DOESN'T WORK ON MAC OS X!
		}
	}
}