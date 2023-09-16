package main

import (
	"sync"

	"github.com/rtfmkiesel/msolsprayng/pkg/logger"
	"github.com/rtfmkiesel/msolsprayng/pkg/options"
	"github.com/rtfmkiesel/msolsprayng/pkg/result"
	"github.com/rtfmkiesel/msolsprayng/pkg/sprayer"
)

func main() {
	opt, err := options.Parse()
	if err != nil {
		logger.Critical(err.Error())
	}

	chanJobs := make(chan string)
	wgSprayer := new(sync.WaitGroup)
	chanResults := make(chan result.Result)
	wgResults := new(sync.WaitGroup)

	go result.Runner(wgResults, chanResults, opt)
	wgResults.Add(1)

	for i := 0; i < opt.Workers; i++ {
		go sprayer.Runner(wgSprayer, chanJobs, chanResults, opt.Password)
		wgSprayer.Add(1)
	}

	for _, user := range opt.Users {
		chanJobs <- user
	}

	close(chanJobs)
	wgSprayer.Wait()

	close(chanResults)
	wgResults.Wait()
}
