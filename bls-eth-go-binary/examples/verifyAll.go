package main

import (
	// "context"
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"

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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

	// Only consider first 1.75mm
	files := []*drive.File{}
	for _, file := range allFiles {
		blockSlot, _ := strconv.ParseInt(strings.Split(strings.Split(file.Name, ".")[0], "_")[2], 10, 64)
		if blockSlot > 1750000 {
			break
		}
		files = append(files, file)
	}
	
	var outWg sync.WaitGroup
	for _, f := range files {
			outWg.Add(1)
			go func (f *drive.File) {
				reader, err := readFile(srv, f.Id)
				if err != nil {
					// handle error
				}
				if strings.HasSuffix(f.Name, "csv") {
					csvReader := csv.NewReader(reader)
					records, _ := csvReader.ReadAll()
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
							validators := strtouints(strings.Trim(row[12], "{}"))
	
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
					if err = rows.Close(); err != nil {
						fmt.Println(err)
					}
	
					// write false rows to a file
					writeResultToCSV(falseRows, f.Name)
				}else{
					panic(fmt.Sprintf("get a file with unknown suffix %s", f.Name))
				}
				outWg.Done()
			}(f)	
	}
	outWg.Wait()

}


func writeResultToCSV(rows [][]string, fileName string){
	// Open a file for writing
	output, err := os.Create(fmt.Sprintf("verificationResult/%s.csv", strings.Split(fileName, ".")[0]))
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
	

// func readFile(service *drive.Service, fileId string) (io.Reader, error) {
// 	r, _ := service.Files.Export(fileId, "text/plain").Download()
// 	fmt.Println(r)
// 	return r.Body, nil
// 	// return r.(*drive.MediaData).Body, nil
// }
func readFile(service *drive.Service, fileId string) (io.Reader, error) {
	// Download the file in XLSX format
	r, err := service.Files.Get(fileId).Download()
	if err != nil {
		return nil, err
	}
	return r.Body, nil
}


