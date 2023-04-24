# msolsprayng
This is a Golang port of [github.com/dafthack/MSOLSpray](https://github.com/dafthack/MSOLSpray). This tool will spray one password against a list of Microsoft accounts. Since Microsoft's GraphQL endpoint is very verbose, we get error codes back that can give us information about the login process / account state. Below is a list of error codes. All error codes can be found [here](https://docs.microsoft.com/en-us/azure/active-directory/develop/reference-aadsts-error-codes).

| Code        | Meaning                                                      |
|-------------|--------------------------------------------------------------|
| AADSTS50126 | Invalid Password                                             |
| AADSTS50128 | Tenant not found                                             |
| AADSTS50059 | Tenant not found                                             |
| AADSTS50034 | User does not exist                                          |
| AADSTS50079 | Password correct but MFA present                             |
| AADSTS50076 | Password correct but MFA present                             |
| AADSTS50158 | Password correct but MFA & Conditional Access Policy present |
| AADSTS53003 | Password correct but Conditional Access Policy present       |
| AADSTS50053 | Account locked                                               |
| AADSTS50057 | Account disabled                                             |
| AADSTS50055 | Password correct but expired                                 |

## Installation
### Binaries
Download the pre built binaries [here](https://github.com/rtfmkiesel/msolsprayng/releases).

### With go
```bash
go install github.com/rtfmkiesel/msolsprayng@latest
```

## Build from source
```bash
git clone https://github.com/rtfmkiesel/msolsprayng
cd msolsprayng
# to build binary in the current directory
go build -ldflags="-s -w" .
# to build & install binary into GOPATH/bin
go install .
```
## Usage
```
msolsprayng -u users.txt -P "Summer2023" [OPTIONS]

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
```

## Legal
**I'm not responsible if your IP address get blocked by Microsoft.** Additionally, this code is provided for educational use only. If you engage in any illegal activity the author does not take any responsibility for it. By using this code, you agree with these terms.

