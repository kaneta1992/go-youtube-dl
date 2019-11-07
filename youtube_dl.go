package youtube_dl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
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

func monitorStdout(reader io.Reader) {
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
	for scanner.Scan() {
		fmt.Printf("\r%s", scanner.Text())
	}
}

func runCommand(cmd *exec.Cmd) error {
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go monitorStdout(outReader)

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (y *YoutubeDl) Download(url string) error {
	options := &youtubeDlOptions{}
	y.addUserOptionIfNeeded(options)
	options.add("--format", "'bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best'")
	command := fmt.Sprintf("%s %s %s", commandName, options.toString(), url)
	fmt.Println(command)
	cmd := exec.Command("bash", "-c", command)
	return runCommand(cmd)
}

func (y *YoutubeDl) DownloadAudio(url, format string) error {
	options := &youtubeDlOptions{}
	y.addUserOptionIfNeeded(options)
	options.add("--format", "'bestaudio'")
	options.add("--audio-format", format)
	options.add("--audio-quality", "0")
	options.add("--extract-audio", "")
	command := fmt.Sprintf("%s %s %s", commandName, options.toString(), url)
	fmt.Println(command)
	cmd := exec.Command("bash", "-c", command)
	return runCommand(cmd)
}
