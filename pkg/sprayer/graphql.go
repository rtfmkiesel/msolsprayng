package sprayer

import (
	"encoding/json"
	"fmt"
	"strings"
)

type graphQLError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorCodes       []int  `json:"error_codes"`
	Timestamp        string `json:"timestamp"`
	TraceID          string `json:"trace_id"`
	CorrelationID    string `json:"correlation_id"`
	ErrorURI         string `json:"error_uri"`
}

type microsoftError struct {
	Code string
	Msg  string
}

// lookupErrorCode() returns the meaning of a response from the MS GraphQL API
func lookupErrorCode(responseBody []byte) (microsoftError, error) {
	var response = string(responseBody)

	// Taken from the original repo https://github.com/dafthack/MSOLSpray
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/reference-aadsts-error-codes
	if strings.Contains(response, "AADSTS50126") {
		return microsoftError{
			Code: "AADSTS50126",
			Msg:  "Invalid Password",
		}, nil

	} else if strings.Contains(response, "AADSTS50128") {
		return microsoftError{
			Code: "AADSTS50128",
			Msg:  "Tenant not found",
		}, nil

	} else if strings.Contains(response, "AADSTS50059") {
		return microsoftError{
			Code: "AADSTS50059",
			Msg:  "Tenant not found",
		}, nil

	} else if strings.Contains(response, "AADSTS50034") {
		return microsoftError{
			Code: "AADSTS50034",
			Msg:  "User does not exist",
		}, nil

	} else if strings.Contains(response, "AADSTS50079") {
		return microsoftError{
			Code: "AADSTS50079",
			Msg:  "Password correct but MFA present",
		}, nil

	} else if strings.Contains(response, "AADSTS50076") {
		return microsoftError{
			Code: "AADSTS50076",
			Msg:  "Password correct but MFA present",
		}, nil

	} else if strings.Contains(response, "AADSTS50158") {
		return microsoftError{
			Code: "AADSTS50158",
			Msg:  "Password correct but MFA & Conditional Access Policy present",
		}, nil

	} else if strings.Contains(response, "AADSTS53003") {
		// Thanks to https://github.com/mgeeky
		return microsoftError{
			Code: "AADSTS53003",
			Msg:  "Password correct but Conditional Access Policy present",
		}, nil

	} else if strings.Contains(response, "AADSTS50053") {
		accountsLocked++
		return microsoftError{
			Code: "AADSTS50053",
			Msg:  "Account locked",
		}, nil

	} else if strings.Contains(response, "AADSTS50057") {
		return microsoftError{
			Code: "AADSTS50057",
			Msg:  "Account disabled",
		}, nil

	} else if strings.Contains(response, "AADSTS50055") {
		return microsoftError{
			Code: "AADSTS50055",
			Msg:  "Password correct but expired",
		}, nil
	}

	// Got unknown error, unmarshal JSON response
	var jsonResponse graphQLError
	err := json.Unmarshal([]byte(response), &jsonResponse)
	if err != nil {
		return microsoftError{}, err
	}

	return microsoftError{
		// Return the first error code and the first line of the error message
		Code: fmt.Sprintf("AADSTS%d", jsonResponse.ErrorCodes[0]),
		Msg:  strings.Split(jsonResponse.ErrorDescription, "\r\n")[0],
	}, nil
}
