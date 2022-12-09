package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kzgceremony"
	"kzgceremony/client"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

type Config struct {
	sequencerURL string
	randomness   string
	sleepTime    uint64
}

func main() {
	fmt.Println("eth-kzg-ceremony-alt")
	fmt.Printf("====================\n\n")

	red := color.New(color.FgRed)
	redB := color.New(color.FgRed, color.Bold)
	cyan := color.New(color.FgCyan)
	cyanB := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgHiGreen)
	greenB := color.New(color.FgHiGreen, color.Bold)

	config := Config{}
	flag.StringVarP(&config.sequencerURL, "url", "u",
		"https://kzg-ceremony-sequencer-dev.fly.dev", "sequencer url")
	flag.StringVarP(&config.randomness, "rand", "r",
		"", "randomness")
	flag.Uint64VarP(&config.sleepTime, "sleeptime", "s",
		10, "time (seconds) sleeping before trying again to be the next contributor")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	c := client.NewClient(config.sequencerURL)

	// get status
	msgStatus, err := c.GetCurrentStatus()
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
	fmt.Println(msgStatus)

	if config.randomness == "" {
		cyanB.Println("To contribute to the ceremony, please set your randomness. Use -h to show the available flags.")
		os.Exit(0)
	}

	if len([]byte(config.randomness)) < kzgceremony.MinRandomnessLen {
		redB.Printf("Randomness must be longer than %d, current length: %d\n",
			kzgceremony.MinRandomnessLen, len([]byte(config.randomness)))
		os.Exit(1)
	}

	// Auth
	msgReqLink, err := c.GetRequestLink()
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}

	green.Printf("Please go to\n%s\n and authenticate with Github.\n", msgReqLink.GithubAuthURL)
	fmt.Println("(currently only Github auth is supported)")

	greenB.Printf("Paste here the RawData from the auth answer:\n")
	s, err := readInput()
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
	var authMsg client.MsgAuthCallback
	if err = json.Unmarshal([]byte(s), &authMsg); err != nil {
		red.Println(err)
		os.Exit(1)
	}
	fmt.Print("Parsed auth msg: ")
	cyan.Printf("%#v\n", authMsg)

	// TODO this will be only triggered by a flag
	// msg, err := c.PostAbortContribution(authMsg.SessionID)
	// if err != nil {
	//         red.Println(err)
	//	   os.Exit(1)
	// }
	// fmt.Println("ABORT", string(msg))
	// os.Exit(0)

	// Get on queue
	var prevBatchContribution *kzgceremony.BatchContribution
	for {
		fmt.Printf("%s sending try_contribute\n", time.Now().Format("2006-01-02 15:04:05"))
		var retry bool
		prevBatchContribution, retry, err = c.PostTryContribute(authMsg.SessionID)
		if err != nil {
			red.Println(err)
		}
		if !retry {
			break
		}
		fmt.Printf("%s try_contribute unsuccessful, going to sleep %d seconds\n",
			time.Now().Format("2006-01-02 15:04:05"), config.sleepTime)
		time.Sleep(time.Duration(config.sleepTime) * time.Second)
	}

	// get latest state
	// currentState, err := c.GetCurrentState()
	// if err != nil {
	//         red.Println(err)
	//	   os.Exit(1)
	// }

	fmt.Println("starting to compute new contribution")
	newBatchContribution, err := prevBatchContribution.Contribute([]byte(config.randomness))
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
	// store contribution
	fmt.Println("storing contribution.json")
	b, err := json.Marshal(newBatchContribution)
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile("contribution.json", b, 0600)
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}

	// send contribution
	fmt.Println("sending contribution")
	receipt, err := c.PostContribute(authMsg.SessionID, newBatchContribution)
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
	fmt.Println("Receipt:")
	green.Println(receipt)

	// store receipt
	fmt.Println("storing contribution_receipt.json")
	b, err = json.Marshal(receipt)
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile("contribution_receipt.json", b, 0600)
	if err != nil {
		red.Println(err)
		os.Exit(1)
	}
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// remove the delimeter from the string
	input = strings.TrimSuffix(input, "\n")
	return input, nil
}
