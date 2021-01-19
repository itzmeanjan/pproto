package main

import (
	"log"
	"time"
)

func main() {
	dataSeq := "data_seq.bin"
	count := 1000000

	start := time.Now()
	ret := SequentialWriteToFile(dataSeq, count)
	end := time.Now()

	if ret {
		log.Printf("[+] Wrote %d protocol buffer encoded entries in %s [ Sequential ]\n", count, end.Sub(start))
	}

}
