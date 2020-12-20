package main

// @raifpy - Mon Dec 21 02:18:27 +03 2020 | go 1.16.5

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/tucnak/telebot.v2"
)

var globalAdminID int

func main() {
	botToken := os.Getenv("token")
	adminIDString := os.Getenv("adminid")

	if botToken == "" || adminIDString == "" {
		fmt.Fprintln(os.Stderr, "'token' or 'adminid' env empty.\nUsage: env token=yourTelegramBotToken fileid=yourUserID "+os.Args[0])
		os.Exit(1)
	}

	adminID, err := strconv.Atoi(adminIDString)
	if err != nil {
		fmt.Fprintln(os.Stderr, "adminid must be integer!")
		os.Exit(1)
	}

	globalAdminID = adminID // init

	bot, err := telebot.NewBot(telebot.Settings{Token: botToken, Verbose: true, Poller: &telebot.LongPoller{Timeout: 10 * time.Second}})
	if err != nil {
		log.Fatalln(err)
		return
	}

	bot.Handle("/start", func(m *telebot.Message) {
		bot.Send(m.Sender, "`Hi. I am bridge for Server & Telegram.`\nBasicly i can upload files for yourself.\nVisit my source code here: github.com/raifpy/tgServerBridge", telebot.ModeMarkdown)
	})

	bot.Handle("/help", func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}

		bot.Send(m.Sender, "/help `show this message`\n/ls `listdir`\n/cd <dir> `chance dir`\n/pwd `show 'where i am'`\n/exec <command> <param> .. `Run command on server`\n<any> get file/list dir", telebot.ModeMarkdown)
	})

	bot.Handle("/cd", func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}
		spi := strings.Split(m.Text, " ")
		if len(spi) == 1 {
			_, err := bot.Send(m.Sender, "Usage: `/cd dirName`", telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
		err := os.Chdir(spi[1])
		if err != nil {
			_, err := bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
	})

	bot.Handle("/pwd", func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}
		pwd, err := os.Getwd()
		if err != nil {
			_, err = bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
		_, err = bot.Send(m.Sender, "`"+pwd+"`", telebot.ModeMarkdown)
		checkMessageSended(bot, err)
	})

	bot.Handle("/ls", func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}
		_, err := bot.Send(m.Sender, ls("."), telebot.ModeMarkdown)
		checkMessageSended(bot, err)

	})

	bot.Handle("/exec", func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}
		spi := strings.Split(m.Text, " ")
		if len(spi) == 1 {
			_, err := bot.Send(m.Sender, "Usage: `/exec command param param2 ...`", telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
		var cmd *exec.Cmd
		if len(spi) > 2 {
			cmd = exec.Command(spi[1], spi[2:]...)
		} else {
			cmd = exec.Command(spi[1])
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			_, err := bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			_, err := bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
		stdErrBuf := bufio.NewScanner(stderr)
		stdErrBuf.Split(bufio.ScanWords)
		stdOutBuf := bufio.NewScanner(stdout)
		stdOutBuf.Split(bufio.ScanWords)

		go func() {
			for stdErrBuf.Scan() {
				_, err := bot.Send(m.Sender, stdErrBuf.Text(), telebot.ModeMarkdown)
				checkMessageSended(bot, err)
			}
		}()
		go func() {
			for stdOutBuf.Scan() {
				_, err := bot.Send(m.Sender, stdOutBuf.Text(), telebot.ModeMarkdown)
				checkMessageSended(bot, err)
			}
		}()

		cmd.Start()

		//buf := bufio.Ne
	})

	bot.Handle(telebot.OnText, func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}
		fileinfo, err := os.Stat(m.Text)
		if err != nil {
			_, err := bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}
		if fileinfo.IsDir() {

			_, err := bot.Send(m.Sender, ls(m.Text), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
			return
		}

		_, err = bot.Send(m.Sender, &telebot.Document{File: telebot.FromDisk(m.Text), FileName: m.Text})
		if err != nil {
			_, err := bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)
		}
	})

	bot.Handle(telebot.OnDocument, func(m *telebot.Message) {
		if m.Sender.ID != adminID {
			return
		}
		err := bot.Download(m.Document.MediaFile(), m.Document.FileName)
		if err != nil {
			_, err := bot.Send(m.Sender, formatErr(err), telebot.ModeMarkdown)
			checkMessageSended(bot, err)

		}

	})

	bot.Start()
}

func getTypeDir(ok bool) string {
	if ok {
		return "dir"
	}
	return "file"
}

func ls(dirName string) string {
	dir, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err.Error()
	}
	var str = "List: `" + filepath.Dir(dirName) + "`\n"
	for _, d := range dir {
		str += fmt.Sprintf("`%s` - %s `%dkb` %s\n", d.Name(), d.Mode().String(), (d.Size() / 1024), getTypeDir(d.IsDir()))
	}
	return str
}

func formatErr(err error) string {
	return "Error: `" + err.Error() + "`"
}

func checkMessageSended(bot *telebot.Bot, err error) {
	if err != nil {
		log.Println(err.Error())
		bot.Send(&telebot.User{ID: globalAdminID}, "Error on sendMessage: `"+err.Error()+"`", telebot.ModeMarkdown)
	}
}
