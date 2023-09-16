package options

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rtfmkiesel/msolsprayng/pkg/logger"
)

type Options struct {
	Users    []string
	Password string
	JSON     bool
	Workers  int
}

// Parse() returns the Options needed for the rest of the programm to function
func Parse() (opt Options, err error) {
	opt = Options{}

	var flagUserfile string
	var flagPasswordFile string
	flag.StringVar(&flagUserfile, "u", "", "")
	flag.StringVar(&flagUserfile, "users", "", "")
	flag.StringVar(&flagPasswordFile, "p", "", "")
	flag.StringVar(&flagPasswordFile, "password-file", "", "")
	flag.StringVar(&opt.Password, "P", "", "")
	flag.StringVar(&opt.Password, "Password", "", "")
	flag.StringVar(&logger.Logfile, "o", "", "")
	flag.StringVar(&logger.Logfile, "outfile", "", "")
	flag.BoolVar(&opt.JSON, "j", false, "")
	flag.BoolVar(&opt.JSON, "json", false, "")
	flag.BoolVar(&logger.Verbose, "v", false, "")
	flag.BoolVar(&logger.Verbose, "verbose", false, "")
	flag.IntVar(&opt.Workers, "w", 1, "")
	flag.IntVar(&opt.Workers, "workers", 1, "")
	flag.Usage = func() { usage() }
	flag.Parse()

	if flagUserfile == "" {
		return opt, fmt.Errorf("Missing argument '-u'")
	} else if opt.Password != "" && flagPasswordFile != "" {
		return opt, fmt.Errorf("Only one password argument ('-p' or '-P') required")
	} else if opt.Password == "" && flagPasswordFile == "" {
		return opt, fmt.Errorf("Missing argument '-p' or '-P'")
	} else if opt.Workers > 3 {
		logger.Info("Max value for '-w' is 3. Reducing %d to 3", opt.Workers)
		opt.Workers = 3
	} else if opt.JSON {
		// If JSON is wanted, verbose needs to be disabled
		logger.Verbose = false
	}

	// Load password from file
	if flagPasswordFile != "" {
		_, err := os.Stat(flagPasswordFile)
		if os.IsNotExist(err) {
			return opt, fmt.Errorf("File '%s' does not exist", flagPasswordFile)
		}

		file, err := os.Open(flagPasswordFile)
		if err != nil {
			return opt, fmt.Errorf("Could not open file '%s': %s", flagPasswordFile, err.Error())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		scanner.Scan()
		opt.Password = scanner.Text() // Only first line

		logger.Info("Using password '%s' from file", opt.Password)
	}

	// Load input file
	_, err = os.Stat(flagUserfile)
	if os.IsNotExist(err) {
		return opt, fmt.Errorf("File '%s' does not exist", flagUserfile)
	}

	file, err := os.Open(flagUserfile)
	if err != nil {
		return opt, fmt.Errorf("Could not open file '%s': %s", flagUserfile, err.Error())
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		// Normalize
		user := strings.ToLower(s.Text())

		// Only process if email is new
		if !contains(opt.Users, user) {
			// Only append if valid email
			if isEmail(user) {
				opt.Users = append(opt.Users, user)
			} else {
				logger.Info("Ignoring '%s', invalid email", user)
			}
		} else {
			logger.Info("Ignoring '%s', duplicate", user)
		}
	}

	if len(opt.Users) == 0 {
		return opt, fmt.Errorf("No valid emails supplied")
	} else {
		logger.Info("Spraying %d users", len(opt.Users))
	}

	return opt, nil
}

func usage() {
	fmt.Printf(`Usage:
    msolsprayng [OPTIONS]
	
Options:
    -u, --users              <string>    Path to file containing E-Mail addresses to be sprayed
    -p, --password-file      <string>    Path to file containing the password to spray (will use first line)
    -P, --Password           <string>    Password to spray (argument)

    -o, --outfile            <string>    Path to the logfile
    -j, --json                           Format output as JSON (default: false)
    -v, --verbose                        Enable verbose output (default: false)

    -w, --workers            <int>       Amount of workers / "threads" (default: 1, max 3)
    -h, --help                           Prints this text

`)
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

// isEmail() will return true is a email is valid
//
// https://github.com/asaskevich/govalidator/blob/master/patterns.go#L7
func isEmail(email string) bool {
	emailRegex := regexp.MustCompile("^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

	return emailRegex.MatchString(email)
}
