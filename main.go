package main

import (
	"log"
	"time"
)

func main() {
	dataSeq := "data_seq.bin"
	dataCon := "data_con.bin"
	count := 10000000

	start := time.Now()
	ret := WriteAllToFile(dataSeq, count)
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

}
