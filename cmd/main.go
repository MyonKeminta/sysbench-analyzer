package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/MyonKeminta/sysbench-analyzer/lib"
)

func main() {
	// TODO: Process command line args more elegantly
	// TODO: --help
	var command string
	if len(os.Args) <= 1 {
		command = "check"
	} else {
		command = os.Args[1]
	}

	args := os.Args[2:]

	switch command {
	case "check":
		check(args)
		break
	case "plot":
		plot(args)
		break
	case "long-check":
		longCheck(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %v\nUsage: %v {check|plot} [args...]\n", command, os.Args[0])
	}

}

func check(args []string) {
	fmt.Printf("Check from stdin for QPS dropping, threshold 70%%, ignoring invalid lines...\n")

	success, violations := lib.CheckQPSDropFromSysbenchOutputStream(os.Stdin, 0.7)

	if success {
		fmt.Println("Passed.")
	} else {
		fmt.Println("Abnormal Lines:")
		fmt.Print(violations)
	}
}

func plot(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Missing file Name.\nUsage: %v plot <output-img-file>", os.Args[0])
	}
	imageFile := args[0]

	analyzer := lib.NewSysbenchAnalyzer([]lib.Rule{}, true)
	records, _ := analyzer.ParseToEnd(os.Stdin)

	err := lib.PlotQPS(records, imageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Plotting failed. Error:\n%v\n", err)
	}
}

type longCheckConfig struct {
	Tag             string
	SlackToken      string
	SlackChannel    string
	SysbenchCmdLine string
	CheckLines      int
	QpsLowerLimit   float64
	QpsUpperLimit   float64
	Loop            bool
}

func defaultConfig() longCheckConfig {
	return longCheckConfig{
		CheckLines:    1000,
		QpsLowerLimit: 0.7,
		QpsUpperLimit: 1.3,
	}
}

type slackNotifier struct {
	tag          string
	slackToken   string
	slackChannel string

	lastNotifyTime time.Time
	messages       []string
	messageChan    chan string
}

func newSlackNotifier(tag string, slackToken string, slackChannel string) *slackNotifier {
	return &slackNotifier{
		tag:          tag,
		slackToken:   slackToken,
		slackChannel: slackChannel,
		messageChan:  make(chan string, 100),
	}
}

func (s *slackNotifier) start(ctx context.Context) {
	go s.run(ctx)
}

func (s *slackNotifier) run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	select {
	case msg := <-s.messageChan:
		s.messages = append(s.messages, msg)
		if time.Since(s.lastNotifyTime) > time.Minute*5 {
			s.flushToSlack()
		}

	case <-ticker.C:
		if time.Since(s.lastNotifyTime) > time.Minute*5 {
			s.flushToSlack()
		}

	case <-ctx.Done():
		s.flushToSlack()
		return
	}
}

func (s *slackNotifier) flushToSlack() {
	if len(s.messages) == 0 {
		return
	}

	if len(s.slackToken) == 0 || len(s.slackChannel) == 0 {
		return
	}

	client := slack.New(s.slackToken)
	msgToSlack := "*Alert from " + s.tag + "*:\n\n"
	for i, msg := range s.messages {
		if i > 3 {
			msgToSlack += fmt.Sprintf("...more %v alerts", len(s.messages)-3)
			break
		}
		msgToSlack += msg + "\n\n"
	}

	msgToSlack = strings.TrimSpace(msgToSlack)

	_, _, err := client.PostMessage(s.slackChannel, slack.MsgOptionText(msgToSlack, false))
	if err != nil {
		log.Printf("[Error] Failed sending message to slack: %v", err)
	}
	s.lastNotifyTime = time.Now()
}

func (s *slackNotifier) put(msg string) {
	log.Println("[to slack] " + msg)
	s.messageChan <- "* " + time.Now().String() + ":\n" + msg
}

func longCheck(args []string) {
	if len(args) == 0 {
		log.Fatalf("Please specify config file.\nUsage:\n  ./sbanalyzer long-check <config-file>")
		os.Exit(-1)
	}

	cfg := defaultConfig()
	cfgFile, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatalf("Failed loading config file \"%v\": %v", args[0], err)
	}

	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Fatalf("Failed parsing config file \"%v\": %v", args[0], err)
	}
	log.Printf("Config: %+v", cfg)

	if len(cfg.SlackToken) == 0 || len(cfg.SlackChannel) == 0 {
		log.Printf("Warning: Slack Alerting not configured. No alret will be sent.")
	}

	analyzer := lib.NewSysbenchAnalyzer([]lib.Rule{lib.NewQPSDropRule(cfg.QpsLowerLimit)}, false)

	notifier := newSlackNotifier(cfg.Tag, cfg.SlackToken, cfg.SlackChannel)
	ctx, cancel := context.WithCancel(context.Background())
	notifier.start(ctx)
	defer cancel()

	for {
		log.Printf("Running command: /bin/sh -c %v", cfg.SysbenchCmdLine)
		cmd := exec.Command("/bin/sh", "-c", cfg.SysbenchCmdLine)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			notifier.put(fmt.Sprintf("Failed getting StdoutPipe from command, err: %v", err))
			time.Sleep(time.Duration(time.Minute * 3))
			continue
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			notifier.put(fmt.Sprintf("Failed getting StderrPipe from command, err: %v", err))
			time.Sleep(time.Duration(time.Minute * 3))
			continue
		}

		err = cmd.Start()
		if err != nil {
			notifier.put(fmt.Sprintf("Failed starting the process, err: %v", err))
			time.Sleep(time.Duration(time.Minute * 3))
			continue
		}
		log.Printf("Process started")

		scanner := bufio.NewScanner(stdout)
		outputChan := make(chan string)
		go func() {
			for {
				scanRes := scanner.Scan()
				if !scanRes {
					log.Printf("Output reached EOF")
				}

				str := scanner.Text()
				if len(str) > 0 {
					log.Println("[stdout] " + str)
					outputChan <- str
				}
				if !scanRes {
					break
				}
			}
			outputChan <- "__EOF__"
		}()

		errReaderStopChan := make(chan int)
		stderrScanner := bufio.NewScanner(stderr)
		go func() {
			for {
				scanRes := stderrScanner.Scan()

				str := stderrScanner.Text()
				if len(str) > 0 {
					log.Println("[stderr] " + str)
					notifier.put(fmt.Sprintf("Output from stderr:\n```\n%v\n```", str))
				}
				if !scanRes {
					break
				}
			}
			errReaderStopChan <- 0
		}()

		hasValidOutput := false

		var records []lib.Record
		invalidText := ""
		var lastReadTime time.Time
		isStopped := false

		previousRecordTime := int32(0)

		for {
			if len(invalidText) > 0 && (time.Since(lastReadTime) >= time.Second*2 || isStopped) {
				notifier.put(fmt.Sprintf("Unexpected output: \n```\n%v\n```", invalidText))
			}

			if isStopped {
				break
			}

			line := <-outputChan

			if strings.HasPrefix(line, "SQL statistics") {
				log.Printf("Sysbench finished")
				records = []lib.Record{}
				for {
					remainingLine := <-outputChan
					if remainingLine == "__EOF__" {
						line = remainingLine
						break
					}
					if _, err := lib.ParseRecord(remainingLine); err == nil {
						line = remainingLine
						break
					}
				}
			}

			if line == "__EOF__" {
				if !hasValidOutput {
					notifier.put("Processed exited but no output can be parsed as sysbench's report")
				}
				isStopped = true
				continue
			}

			record, err := lib.ParseRecord(line)
			if err != nil {
				if len(records) != 0 {
					if len(invalidText) > 0 {
						invalidText += "\n"
					}
					invalidText += line
				}
				continue
			}

			hasValidOutput = true
			if record.Second <= previousRecordTime {
				log.Printf("Seems that sysbench has restarted here")
				records = []lib.Record{}
			}
			previousRecordTime = record.Second

			records = append(records, record)

			if len(records) >= cfg.CheckLines || (len(records) > 0 && isStopped) {
				violations := analyzer.AnalyzeParsedRecords(records)
				if len(violations) > 0 {
					violationText := ""
					for i, v := range violations {
						if i == 10 {
							violationText += fmt.Sprintf("...more %v items\n", len(violations)-10)
							break
						}
						violationText += v.RecordText + "\n  *" + v.Description + "\n"
					}
					notifier.put(fmt.Sprintf("Check failed:\n```\n%v```", violationText))
				}
				// Pop out half of the records
				records = records[len(records)/2:]
			}
		}

		<-errReaderStopChan

		err = cmd.Wait()
		if err != nil {
			notifier.put(fmt.Sprintf("Process didn't exits normally: %v", err))
		}

		if !cfg.Loop {
			break
		}
	}
}
