package main

import (
	"log"
	"time"
)

func main() {
	dataSeq := "data_seq.bin"
	dataCon := "data_con.bin"
	compDataSeq := "comp_data_seq.bin"
	count := 10000000

	start := time.Now()
	ret := SequentialWriteToFile(dataSeq, count)
	end := time.Now()

	if ret {
		log.Printf("[+] Wrote %d protocol buffer encoded entries in %s [ Sequential ]\n", count, end.Sub(start))
	}

	start = time.Now()
	ret = ConcurrentWriteAllToFile(dataCon, count)
	end = time.Now()

	if ret {
		log.Printf("[+] Wrote %d protocol buffer encoded entries in %s [ Concurrent ]\n", count, end.Sub(start))
	}

	start = time.Now()
	ret, _count := SequentialReadFromFile(dataSeq)
	end = time.Now()

	if ret {
		log.Printf("[+] Read %d protocol buffer encoded entries in %s [ Sequential ]\n", _count, end.Sub(start))
	}

	start = time.Now()
	ret, _count = ConcurrentReadFromFile(dataCon)
	end = time.Now()

	if ret {
		log.Printf("[+] Read %d protocol buffer encoded entries in %s [ Concurrent ]\n", _count, end.Sub(start))
	}

	start = time.Now()
	ret = CompressedSequentialWriteToFile(compDataSeq, count)
	end = time.Now()

	if ret {
		log.Printf("[+] Wrote %d protocol buffer encoded entries in %s [ Sequential + Compressed ]\n", count, end.Sub(start))
	}
}
