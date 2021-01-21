package main

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/itzmeanjan/pproto/pb"
	"google.golang.org/protobuf/proto"
)

// SequentialReadFromFile - Given path to protocol buffer encoded data file
// attempting to read deserialised content of file in-memory, sequentially
// but in buffer fashion, so that memory footprint stays low
func SequentialReadFromFile(file string) (bool, int) {

	// Opening file in read only mode
	fd, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {

		log.Printf("[!] Error : %s\n", err.Error())
		return false, 0

	}

	// file handle to be closed when whole file is read i.e.
	// EOF reached
	defer fd.Close()

	// count of entries read back from file
	var count int

	for {

		buf := make([]byte, 4)

		// reading size of next protocol buffer encoded
		// data chunk
		if _, err := fd.Read(buf); err != nil {

			// reached EOF, good to get out of loop
			if err == io.EOF {
				break
			}

			log.Printf("[!] Error : %s\n", err.Error())
			return false, count

		}

		// converting size of next data chunk to `uint`
		// so that memory allocation can be performed
		// for next read
		size := binary.LittleEndian.Uint32(buf)

		// allocated buffer where to read next protocol buffer
		// serialized data chunk
		data := make([]byte, size)

		if _, err = fd.Read(data); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			return false, count

		}

		// attempting to deserialize protocol buffer encoded
		// data into something meaningful
		cpu := &pb.CPU{}
		if err := proto.Unmarshal(data, cpu); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			return false, count

		}

		count++
	}

	return true, count
}

// ConcurrentReadFromFile - Reading content from file and letting
// workers process those content concurrently
func ConcurrentReadFromFile(file string) (bool, int) {

	// Opening file in read only mode
	fd, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {

		log.Printf("[!] Error : %s\n", err.Error())
		return false, 0

	}

	// file handle to be closed when whole file is read i.e.
	// EOF reached
	defer fd.Close()

	// count of entries read back from file
	var count uint64

	control := make(chan bool, 1000)
	entryCount := make(chan uint64)
	done := make(chan bool)

	go UnmarshalCoordinator(control, entryCount, done)

	for {

		buf := make([]byte, 4)

		// reading size of next protocol buffer encoded
		// data chunk
		if _, err := fd.Read(buf); err != nil {

			// reached EOF, good to get out of loop
			if err == io.EOF {
				break
			}

			log.Printf("[!] Error : %s\n", err.Error())
			break

		}

		// converting size of next data chunk to `uint`
		// so that memory allocation can be performed
		// for next read
		size := binary.LittleEndian.Uint32(buf)

		// allocated buffer where to read next protocol buffer
		// serialized data chunk
		data := make([]byte, size)

		if _, err = fd.Read(data); err != nil {

			log.Printf("[!] Error : %s\n", err.Error())
			break

		}

		count++
		go UnmarshalData(data, control)

	}

	// letting coordinator know that `count` many workers
	// should let it know about their respective status of job
	entryCount <- count
	// waiting for coordinator to let us know
	// that all workers have completed their job
	<-done

	return true, int(count)

}

// UnmarshalData - Given byte array read from file, attempting
// to unmarshall it into structured data, with synthetic delay
//
// Also letting coordinator go routine know that this worker
// has completed its job
func UnmarshalData(data []byte, control chan bool) {

	// synthetic delay to emulate database interaction
	time.Sleep(time.Duration(rand.Intn(400)+100) * time.Microsecond)

	cpu := &pb.CPU{}
	if err := proto.Unmarshal(data, cpu); err != nil {

		log.Printf("[!] Error : %s\n", err.Error())
		control <- false
		return

	}

	control <- true

}

// UnmarshalCoordinator - Given a lot of unmarshal workers to be
// created for processing i.e. deserialize & put into DB, more entries
// in smaller amount of time, they need to be synchronized properly
//
// That's all this go routine does
func UnmarshalCoordinator(control <-chan bool, count <-chan uint64, done chan bool) {

	// letting main go routine know reading & processing all entries
	// done successfully, while getting out of this execution context
	defer func() {
		done <- true
	}()

	// received processing done count from worker go routines
	var success uint64
	// how many workers should actually let this go routine know
	// their status i.e. how long is this go routine supposed to
	// wait for all of them to finish
	var expected uint64

	for {
		select {

		case c := <-control:
			if !c {
				log.Fatalf("[!] Error received by unmarshal coordinator\n")
			}

			// some worker just let us know it completed
			// its job successfully
			success++
			// If this satisfies, it's time to exit from loop
			// i.e. all workers have completed their job
			if success == expected {
				return
			}

		// Once reading whole file is done, main go routine
		// knows how many entries are expected, which is to be
		// matched against how many of them actually completed their job
		//
		// Exiting from this loop, that logic is written 👆
		case v := <-count:
			expected = v
		}
	}

}
