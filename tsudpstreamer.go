package main

import (
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"github.com/aristanetworks/goarista/atime" //for getting monotonic clock time
)

const (
	srvAddr = "224.0.0.1:12345"
	maxDatagramSize = 1316
	fName           = "test.ts"
	bitrate         = 1030949
	TSPacketSize    = 188
	fileChunkSize   = 500
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	/*t1 := atime.NanoTime()
	t2 := atime.NanoTime()
	if t1 >= t2 {
		log.Fatalln("t1 should have been strictly less than t2 ", t1, t2)
	} else {
		log.Println("t1 is strictly less than t2 ", t1, t2)
	}
	*/
	ping(srvAddr)
	//serveMulticastUDP(srvAddr, msgHandler)
}

func ping(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}

	packetSize := 7 * TSPacketSize

	c, err := net.DialUDP("udp", nil, addr)
	file, err := os.Open(fName)
	check(err)
	defer file.Close()

	buf := make([]byte, packetSize)
	//reader := bufio.NewReader(file)

	completed := 0
	packetTime := uint64(0)
	timeStart := uint64(0)
	timeStop := uint64(0)
	realTime := uint64(0)

	nanoSleepPacket := syscall.Timespec{}
	nanoSleepPacket.Nsec = 665778 // 1 packet at 100mbps

	timeStart = atime.NanoTime()

	c.SetWriteBuffer(229376)
	for completed != 1 {

		timeStop = atime.NanoTime()
		realTime = (timeStop - timeStart)

		if realTime*bitrate > packetTime*1000000 && completed != 1 {

			tmp, err := file.Read(buf)
			//check(err)
			if err != nil {
				if err == io.EOF {
					completed = 1
					log.Println("HEDEEE")
				} else {
					panic(err)
				}
			}
			if tmp < 0 {
				log.Println("ts sent done.")
				completed = 1
			} else if tmp == 0 {
				completed = 1
			} else {
				log.Println("bytes: ", tmp, ", string: ", hex.Dump(buf[:tmp]))
				c.Write(buf[:tmp])
				packetTime += uint64(packetSize * 8)

			}
		} else {
			log.Println("syscall nanosleep!!!!!!!!!!!!!!!!! ")
			syscall.Nanosleep(&nanoSleepPacket, nil)
		}

		//time.Sleep(1 * time.Second)
	}
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Println(n, "bytes read from", src)
	log.Println(hex.Dump(b[:n]))
}

func serveMulticastUDP(a string, h func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	l.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}
}
