package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	kzgceremony "github.com/arnaucube/eth-kzg-ceremony-alt"
	"github.com/arnaucube/eth-kzg-ceremony-alt/client"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

var (
	redB   = color.New(color.FgRed, color.Bold)
	cyan   = color.New(color.FgCyan)
	cyanB  = color.New(color.FgCyan, color.Bold)
	green  = color.New(color.FgHiGreen)
	greenB = color.New(color.FgHiGreen, color.Bold)
)

func main() {
	fmt.Println("eth-kzg-ceremony-alt")
	fmt.Printf("====================\n")
	fmt.Printf("            https://github.com/arnaucube/eth-kzg-ceremony-alt\n\n")

	var sequencerURL string
	var randomness string
	var sleepTime uint64
	flag.StringVarP(&sequencerURL, "url", "u",
		"https://seq.ceremony.ethereum.org", "sequencer url")
	flag.StringVarP(&randomness, "rand", "r",
		"", fmt.Sprintf("randomness, needs to be bigger than %d bytes", kzgceremony.MinRandomnessLen))
	flag.Uint64VarP(&sleepTime, "sleeptime", "s",
		30, "time (seconds) sleeping before trying again to be the next contributor")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	c := client.NewClient(sequencerURL)

	// get status
	msgStatus, err := c.GetCurrentStatus()
	if err != nil {
		printErrAndExit(err)
	}
	fmt.Println(msgStatus)

	if randomness == "" {
		_, _ =
			cyanB.Println("To contribute to the ceremony, please set your randomness. Use -h to show the available flags.")
		os.Exit(0)
	}

	if len([]byte(randomness)) < kzgceremony.MinRandomnessLen {
		_, _ = redB.Printf("Randomness must be longer than %d, current length: %d\n",
			kzgceremony.MinRandomnessLen, len([]byte(randomness)))
		os.Exit(1)
	}

	// Auth
	fmt.Println("Github Authorization:")
	authMsg := authGH(c)

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
		var status client.Status
		prevBatchContribution, status, err = c.PostTryContribute(authMsg.SessionID)
		if err != nil {
			_, _ = cyan.Println(err)
		}
		if status == client.StatusProceed {
			break
		}
		if status == client.StatusReauth {
			fmt.Println("SessionID has expired, authenticate again with Github:")
			authMsg = authGH(c)
		}
		msgStatus, err := c.GetCurrentStatus()
		if err != nil {
			printErrAndExit(err)
		}
		fmt.Printf("%s try_contribute unsuccessful, lobby size %d, num contrib %d,"+
			"\n    going to sleep %d seconds\n",
			time.Now().Format("2006-01-02 15:04:05"),
			msgStatus.LobbySize, msgStatus.NumContributions,
			sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	// get latest state
	// currentState, err := c.GetCurrentState()
	// if err != nil {
	//         red.Println(err)
	//	   os.Exit(1)
	// }

	fmt.Println("starting to compute new contribution")
	t0 := time.Now()
	newBatchContribution, err := prevBatchContribution.Contribute([]byte(randomness))
	if err != nil {
		printErrAndExit(err)
	}
	fmt.Println("Contribution computed in", time.Since(t0))

	// store contribution
	fmt.Println("storing contribution.json")
	b, err := json.Marshal(newBatchContribution)
	if err != nil {
		printErrAndExit(err)
	}
	err = ioutil.WriteFile("contribution.json", b, 0600)
	if err != nil {
		printErrAndExit(err)
	}

	// send contribution
	fmt.Println("sending contribution")
	receipt, err := c.PostContribute(authMsg.SessionID, newBatchContribution)
	if err != nil {
		printErrAndExit(err)
	}
	fmt.Println("Receipt:")
	_, _ = green.Println(receipt)

	// store receipt
	fmt.Println("storing contribution_receipt.json")
	b, err = json.Marshal(receipt)
	if err != nil {
		printErrAndExit(err)
	}
	err = ioutil.WriteFile("contribution_receipt.json", b, 0600)
	if err != nil {
		printErrAndExit(err)
	}
}

func authGH(c *client.Client) client.MsgAuthCallback {
	msgReqLink, err := c.GetRequestLink()
	if err != nil {
		printErrAndExit(err)
	}

	_, _ = green.Printf("Please go to\n%s\n and authenticate with Github.\n", msgReqLink.GithubAuthURL)
	fmt.Println("(currently only Github auth is supported)")

	_, _ = greenB.Printf("Paste here the RawData from the auth answer:\n")
	s, err := readInput()
	if err != nil {
		printErrAndExit(err)
	}
	var authMsg client.MsgAuthCallback
	if err = json.Unmarshal([]byte(s), &authMsg); err != nil {
		printErrAndExit(err)
	}
	fmt.Print("Parsed auth msg: ")
	_, _ = cyan.Printf("%#v\n", authMsg)
	return authMsg
}

func printErrAndExit(err error) {
	red := color.New(color.FgRed)
	_, _ = red.Println(err)
	os.Exit(1)
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
