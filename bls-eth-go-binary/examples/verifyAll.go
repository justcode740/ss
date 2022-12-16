package main

import (
	// "context"
	// "bufio"
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	// "strconv"
	// "strings"

	"log"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/cenkalti/backoff/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/drive/v3"
)
func t(){
	
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

	// alreadyProcessed := []int{5000, 10000, 15000, 35000}
	// errFiles := 
	file, err := os.Open("largefile.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // scanner := bufio.NewScanner(file)
    // optionally, resize scanner's capacity for lines over 64K, see next example
	errfilesNumber := []int{1505000, 1575000}
    // for scanner.Scan() {
	// 	str:=strings.Split(strings.Split(scanner.Text(), ".")[0], "_")[2]
	// 	val, _ := strconv.Atoi(str)
	// 	errfilesNumber =  append(errfilesNumber, val)
    // }
	fmt.Println(errfilesNumber)

	// ERR beaconchain_attestation_430000.csv
	// ERR beaconchain_attestation_480000.csv
	// ERR beaconchain_attestation_475000.csv
	// ERR beaconchain_attestation_190000.csv
	// ERR beaconchain_attestation_340000.csv
	// ERR beaconchain_attestation_640000.csv
	// ERR beaconchain_attestation_605000.csv
	// ERR beaconchain_attestation_585000.csv
	// ERR beaconchain_attestation_610000.csv
	// strange 84 1FMuyGeOkEvzOEtQEJ6bqO3bkhxRsSYvi Get "https://www.googleapis.com/drive/v3/files/1FMuyGeOkEvzOEtQEJ6bqO3bkhxRsSYvi?alt=media&prettyPrint=false": read tcp 192.168.171.97:51202->142.250.65.234:443: read: connection timed out
	// ERR beaconchain_attestation_645000.csv
	
	// Only consider first 1.75mm 1750000
	files := []*drive.File{}
	for _, file := range allFiles {
		blockSlot, _ := strconv.ParseInt(strings.Split(strings.Split(file.Name, ".")[0], "_")[2], 10, 64)
		// if contains(alreadyProcessed, int(blockSlot)) {continue}
		// if blockSlot < 645000 {continue}
		if blockSlot > 1750000 {		
			break
		}
		if contains(errfilesNumber, int(blockSlot)){
			files = append(files, file)
		}
	}
	fmt.Println(files)
	// nonParallel(files)
	
	var outWg sync.WaitGroup
	sem := semaphore.NewWeighted(1)
	for _, f := range files {
			outWg.Add(1)
			go func (f *drive.File) {
				err := sem.Acquire(context.Background(), 1)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
				reader, err := readFile(srv, f.Id)
				if err != nil {
					// handle error
				}
				if strings.HasSuffix(f.Name, "csv") {
					csvReader := csv.NewReader(reader)
					records, err := csvReader.ReadAll()
					if err != nil{
						fmt.Println("ERR", f.Name)
					}
					falseRows := [][]string{}
					var mutex sync.Mutex
					var wg sync.WaitGroup
					for i := 1; i < len(records); i++{
						// verify att
						row := records[i]
						wg.Add(1)
						go func (row []string) {
							bbr := row[1]
							sig := row[6]
							slot, _ := strconv.ParseUint(row[7], 10, 64)
							cidx, _ := strconv.ParseUint(row[5], 10, 64)
							sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
							sourceRoot := row[9]
							targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
							targetRoot := row[11]
							validators := strtouints(strings.Trim(row[12], "\\[\\]"))
							val := verifyAttestation(bbr, sig, uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot, targetRoot, validators)
							if !val {
								mutex.Lock()
								falseRows = append(falseRows, row)
								mutex.Unlock()
							}
							wg.Done()
						}(row)
					}

					wg.Wait()
					
					// write false rows to a file
					writeResultToCSV(falseRows, f.Name)
					
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
					falseRows := [][]string{}
					var mutex sync.Mutex
					var wg sync.WaitGroup
					for rows.Next() {
						if first {
							first = false
							continue
						}
						row, err := rows.Columns()
						if err != nil {
							fmt.Println(err)
						}
						wg.Add(1)
						go func (row []string) {
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
							
							// fmt.Println(bbr, sig, slot, cidx, sourceEpoch, targetEpoch, sourceRoot, targetRoot, validators)
	
							val := verifyAttestation(bbr, sig, uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot, targetRoot, validators)
							// fmt.Println(val)
							if !val {
								mutex.Lock()
								falseRows = append(falseRows, row)
								mutex.Unlock()
							}
							wg.Done()
						}(row)
					}
					wg.Wait()
					if err = rows.Close(); err != nil {
						fmt.Println(err)
					}
	
					// write false rows to a file
					writeResultToCSV(falseRows, f.Name)
					
				}else{
					panic(fmt.Sprintf("get a file with unknown suffix %s", f.Name))
				}
				sem.Release(1)
				outWg.Done()
			}(f)
	}
	outWg.Wait()
}


func writeResultToCSV(rows [][]string, fileName string){
	newfileName := fmt.Sprintf("verificationResult/%s.csv", strings.Split(fileName, ".")[0])
	if _, err := os.Stat(newfileName); err == nil {
		// if file exist, delete it
		os.Remove(fileName)
	}
	// Open a file for writing
	output, err := os.Create(newfileName)
	
	if err != nil {
		panic(err)
	}
	defer output.Close()

	// Create a CSV writer for the file
	writer := csv.NewWriter(output)

	// Iterate over the array of arrays of strings
	for _, arr := range rows {
		// Write the array of strings to the file as a record (row)
		writer.Write(arr)
	}

	// Flush the buffered data to the file
	writer.Flush()
}

func strtouints(str string) []int{
	// fmt.Println("enter", str)
	strs := strings.Split(str, ",")
	// Convert the strings to uint values
	var ints []int
	for _, s := range strs {
		u, err := strconv.Atoi(strings.Trim(s, " "))
		if err != nil {
			// handle error
		}
		ints = append(ints, int(u))
	}
	return ints
}

func getFiles(srv *drive.Service, folderID string) []*drive.File{
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	pageToken := ""
	files := []*drive.File{}
	for {
		r, err := srv.Files.List().Q(q).Fields("nextPageToken, files(id, name)").PageToken(pageToken).Do()
		if err != nil {
			panic(err)
		}
		files = append(files, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	less := func(i, j int) bool {
		first := files[i].Name
		second := files[j].Name
		val1, _:=strconv.ParseInt(strings.Split(strings.Split(first, ".")[0], "_")[2], 10, 64)
		val2, _:= strconv.ParseInt(strings.Split(strings.Split(second, ".")[0], "_")[2], 10, 64)
		return val1 < val2
	}
	sort.Slice(files, less)
	return files
	
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
			"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
			log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
			log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
			return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
			log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
			tok = getTokenFromWeb(config)
			saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

var times int
	

// func readFile(service *drive.Service, fileId string) (io.Reader, error) {
// 	r, _ := service.Files.Export(fileId, "text/plain").Download()
// 	fmt.Println(r)
// 	return r.Body, nil
// 	// return r.(*drive.MediaData).Body, nil
// }
func readFile(service *drive.Service, fileId string) (io.Reader, error) {
	times++
	// work around googleapi: Error 403: Rate Limit Exceeded, rateLimitExceeded
	if times % 50 == 0 {
		time.Sleep(100 * time.Second)
	}
	var r *http.Response
	// An operation that may fail.
	operation := func() error {
		res, err := service.Files.Get(fileId).Download()
		r = res
		return err // or an error
	}

	// exponential backoff to aovid 503 err like googleapi: got HTTP response code 503 with body: Service Unavailable
	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		// Handle error.
		fmt.Println("strange", times, fileId, err)
	}
	
	return r.Body, nil
}


