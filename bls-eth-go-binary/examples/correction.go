package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)



func readCorrection() map[string]int{
	file, err := os.Open("correction/all2.txt")
    if err != nil {
       
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
	mp := map[string]int{}
    // optionally, resize scanner's capacity for lines over 64K, see next example
    for scanner.Scan() {
        line := scanner.Text()
		splitted := strings.Split(line, ",")
		sig := strings.Trim(splitted[0], " ")
		valIdx, _ := strconv.Atoi(strings.Trim(splitted[1], " "))
		mp[sig]=valIdx

    }

    if err := scanner.Err(); err != nil {
        
    }
	return mp
}