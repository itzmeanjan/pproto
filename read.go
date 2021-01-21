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

// ConcurrentReadFromFile - ...
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
