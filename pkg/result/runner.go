package result

import (
	"encoding/json"
	"sync"

	"github.com/rtfmkiesel/msolsprayng/pkg/logger"
	"github.com/rtfmkiesel/msolsprayng/pkg/options"
)

type Result struct {
	Succsessful bool   `json:"Successful"`
	Email       string `json:"Email"`
	Password    string `json:"Password"`
	ErrorCode   string `json:"ErrorCode"`
	ErrorMsg    string `json:"ErrorMsg"`
}

// result.Runner() will handle the results from the sprayer runner
func Runner(wg *sync.WaitGroup, chanResults <-chan Result, opt options.Options) {
	defer wg.Done()

	for result := range chanResults {
		if opt.JSON {
			jsonResult, err := json.Marshal(result)
			if err != nil {
				logger.Error("Error while converting result to JSON: %s", err.Error())
				continue
			}

			// Pure JSON to terminal
			logger.StdOut(string(jsonResult))

		} else if result.Succsessful {
			logger.Success(result.Email)

		} else if !result.Succsessful {
			logger.Fail(result.Email)
		}
	}
}
