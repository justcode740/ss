package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func verifyDv(){
	// fetch correction
	correction := readCorrection()
	
	// https://docs.google.com/spreadsheets/d/1mDPwQMA1K7nFbRfkBKXsiv1jgI3KTDVz/edit?usp=share_link&ouid=115469787324806160501&rtpof=true&sd=true

	// Read the credentials file
	b, err := ioutil.ReadFile("credentials/credentials-internal.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Parse the credentials
	config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Authenticate and get an access token
	client := getClient(config)

	// Create a new Drive client
	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	// Get the folder ID from the URL
	folderID := "1y9b3rLOrx8vijsrDnk-DmN3dIk9JDFfH"

	// Get all files name and sort 
	allFiles := getFiles(srv, folderID)
	
	// Only consider first 1.75mm 1750000
	files := []*drive.File{}
	for _, file := range allFiles {
		blockSlot, _ := strconv.ParseInt(strings.Split(strings.Split(file.Name, ".")[0], "_")[2], 10, 64)
		if blockSlot > 1750000 {		
			break
		}
		files = append(files, file)
		
	}
	
	for _, f := range files {
		// targetEpoch -> validator idx -> Data
		idxmap := map[int]map[int]RecordData{}
		reader, _ := readFile(srv, f.Id)
		if strings.HasSuffix(f.Name, "csv") {
			csvReader := csv.NewReader(reader)
			records, err := csvReader.ReadAll()
			if err != nil{
				fmt.Println("ERR", f.Name)
			}
			rows := [][]string{}
			// each row is an att
			for i := 1; i < len(records); i++{
				row := records[i]
				bbr := row[1]
				sig := row[6]
				slot, _ := strconv.ParseUint(row[7], 10, 64)
				cidx, _ := strconv.ParseUint(row[5], 10, 64)
				sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
				sourceRoot := row[9]
				targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
				targetRoot := row[11]
				validators := strtouints(strings.Trim(row[12], "\\[\\]"))
				data := Data {
					slot: int(slot),
					cidx: int(cidx),
					bbr: bbr,
					source_epoch: int(sourceEpoch),
					source_root: sourceRoot,
					target_epoch: int(targetEpoch),
					target_root: targetRoot,
				}
				recordData := RecordData {
					data :data,
					validators: validators,
					sig: sig,
				}
				if idx, exist := correction[sig]; exist {
					validators = []int{idx}
				}
				for _, valIdx := range validators {
					if val, exist := idxmap[int(targetEpoch)][valIdx]; exist{
						if val.data != data {
							// double vote detected
							// validx, targetepoch, data1, data2
							// verifyatt for data1
							res1 := verifyAttestation(val.data.bbr, val.sig, uint(val.data.slot), uint(val.data.cidx), uint(val.data.source_epoch), uint(val.data.target_epoch), val.data.source_root, val.data.target_root, val.validators)
							res2 := verifyAttestation(bbr, sig, uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot, targetRoot, validators)
							if res1 && res2 {
								res := []string{
									strconv.FormatInt(int64(valIdx),10), 
									strconv.FormatInt(int64(targetEpoch), 10),
									sig, 
									val.data.String(), 
									data.String(),
								}
								rows = append(rows, res)
							}
							
						}
					}else{
						if idxmap[int(targetEpoch)]==nil{
							idxmap[int(targetEpoch)]=make(map[int]RecordData)
						}
						idxmap[int(targetEpoch)][valIdx] = recordData
					}

				}
				
				
			}
			
			// write false rows to a file
			writeResultToCSV(rows, f.Name)
			
		}else if strings.HasSuffix(f.Name, "xlsx") {
			// Read the file from the byte slice
			file, err := excelize.OpenReader(reader)
			if err != nil {
				// handle error
				panic(err)
			}

			// Print the contents of the first sheet
			rows, err := file.Rows("Sheet1")
			if err != nil {
				// handle error
			}
			// skip first row, titles
			first := true
			ress := [][]string{}
			
			for rows.Next() {
				if first {
					first = false
					continue
				}
				row, err := rows.Columns()
				if err != nil {
					fmt.Println(err)
				}
				
				bbr := row[1]
				sig := row[6]
				slot, _ := strconv.ParseUint(row[7], 10, 64)
				cidx, _ := strconv.ParseUint(row[5], 10, 64)
				sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
				sourceRoot := row[9]
				targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
				targetRoot := row[11]
				// fmt.Println("ow", row[12])
				var validators []int
				if row[12][0]=='{'{
					validators = strtouints(strings.Trim(row[12], "{}"))
				}else if row[12][0]=='['{
					validators = strtouints(strings.Trim(row[12], "\\[\\]"))
				}else{
					fmt.Println(row[12])
					fmt.Println("")
				}

				if idx, exist := correction[sig]; exist {
					validators = []int{idx}
				}

				data := Data {
					slot: int(slot),
					cidx: int(cidx),
					bbr: bbr,
					source_epoch: int(sourceEpoch),
					source_root: sourceRoot,
					target_epoch: int(targetEpoch),
					target_root: targetRoot,
				}
				recordData := RecordData {
					data :data,
					validators: validators,
					sig: sig,
				}
				for _, valIdx := range validators {
					if val, exist := idxmap[int(targetEpoch)][valIdx]; exist{
						if val.data != data {
							// double vote detected
							// validx, targetepoch, data1, data2
							res1 := verifyAttestation(val.data.bbr, val.sig, uint(val.data.slot), uint(val.data.cidx), uint(val.data.source_epoch), uint(val.data.target_epoch), val.data.source_root, val.data.target_root, val.validators)
							res2 := verifyAttestation(bbr, sig, uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot, targetRoot, validators)
							if res1 && res2 {
								res := []string{
									strconv.FormatInt(int64(valIdx),10), 
									strconv.FormatInt(int64(targetEpoch), 10),
									sig, 
									val.data.String(), 
									data.String(),
								}
								ress = append(ress, res)
							}
						}
					}else{
						if idxmap[int(targetEpoch)]==nil{
							idxmap[int(targetEpoch)]=make(map[int]RecordData)
						}
						idxmap[int(targetEpoch)][valIdx] = recordData
					}

				}			
	
			}
			
			if err = rows.Close(); err != nil {
				fmt.Println(err)
			}

			// write false rows to a file
			writeResultToCSV(ress, f.Name)
			
		}else{
			panic(fmt.Sprintf("get a file with unknown suffix %s", f.Name))
		}
	}
	
}