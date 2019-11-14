package youtube_dl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

const commandName = "youtube-dl"

type youtubeDlOptions []string

func (o *youtubeDlOptions) add(option string, args ...string) {
	*o = append(*o, option+" "+strings.Join(args, " "))
}

func (o *youtubeDlOptions) toString() string {
	return strings.Join(*o, " ")
}

type YoutubeDlProgress struct {
	Progress      string
	FileSize      string
	DLSpeed       string
	RemainingTime string
}

type YoutubeDl struct {
	id       string
	password string
}

func Create() *YoutubeDl {
	return &YoutubeDl{}
}

func CreateWithIdAndPassward(id, password string) *YoutubeDl {
	return &YoutubeDl{
		id:       id,
		password: password,
	}
}

func (y *YoutubeDl) addUserOptionIfNeeded(o *youtubeDlOptions) {
	if y.id != "" {
		o.add("--username", y.id)
		o.add("--password", y.password)
	}
}

func monitorStdout(reader io.Reader, ch chan interface{}) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		index := bytes.Index(data, []byte("\r"))
		if index >= 0 {
			return index + 1, data[:index], nil
		}
		if atEOF {
			return len(data) + 1, data, nil
		}
		return 0, nil, nil
	})
	fileNameRegexp := regexp.MustCompile(`\[download\] Destination: (.*)\n`)
	scanner.Scan()

	strs := fileNameRegexp.FindStringSubmatch(scanner.Text())
	//fmt.Printf("%s\n", strs)

	if len(strs) == 0 {
		return
	}

	ch <- strs[1]

	progressRegexp := regexp.MustCompile(`\[download\]\s+(.*)\s+of\s+(.*)\s+at\s+(.*)\s+ETA\s+(.*)`)
	for scanner.Scan() {
		str := scanner.Text()
		strs = progressRegexp.FindStringSubmatch(str)
		if len(strs) == 0 {
			continue
		}
		progress := &YoutubeDlProgress{
			Progress:      strs[1],
			FileSize:      strs[2],
			DLSpeed:       strs[3],
			RemainingTime: strs[4],
		}
		ch <- progress
		//fmt.Printf("\r%s", strs)
		// fmt.Printf("\r%s", scanner.Text())
	}
}

func runCommand(cmd *exec.Cmd, ch chan interface{}) error {
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go monitorStdout(outReader, ch)

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (y *YoutubeDl) createDownloadVideoCommand(url string, simulate bool) *exec.Cmd {
	options := &youtubeDlOptions{}
	y.addUserOptionIfNeeded(options)
	options.add("--format", "'bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best'")
	if simulate {
		options.add("--simulate")
	}
	command := fmt.Sprintf("%s %s %s", commandName, options.toString(), url)
	fmt.Println(command)
	return exec.Command("bash", "-c", command)
}

func (y *YoutubeDl) Download(url string, ch chan interface{}) error {
	cmd := y.createDownloadVideoCommand(url, false)
	err := runCommand(cmd, ch)
	if err != nil {
		ch <- err
		return err
	}
	return nil
}

func (y *YoutubeDl) DownloadSimulate(url string) error {
	cmd := y.createDownloadVideoCommand(url, true)
	return cmd.Run()
}

func (y *YoutubeDl) createDownloadAudioCommand(url, format string, simulate bool) *exec.Cmd {
	options := &youtubeDlOptions{}
	y.addUserOptionIfNeeded(options)
	options.add("--format", "'bestaudio'")
	options.add("--audio-format", format)
	options.add("--audio-quality", "0")
	options.add("--extract-audio", "")
	if simulate {
		options.add("--simulate")
	}
	command := fmt.Sprintf("%s %s %s", commandName, options.toString(), url)
	fmt.Println(command)
	return exec.Command("bash", "-c", command)
}

func (y *YoutubeDl) DownloadAudio(url, format string, ch chan interface{}) error {
	cmd := y.createDownloadAudioCommand(url, format, false)
	err := runCommand(cmd, ch)
	if err != nil {
		ch <- err
		return err
	}
	return nil
}

func (y *YoutubeDl) DownloadAudioSimulate(url, format string) error {
	cmd := y.createDownloadAudioCommand(url, format, true)
	return cmd.Run()
}
