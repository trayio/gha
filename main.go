package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/bgentry/speakeasy"
)

type Authorization struct {
	Scopes []string `json:"scopes"`
	Note   string   `json:"note"`
	// no need for note_url, client_id, client_secret or fingerprint
}

func homeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func main() {
	var user string
	flag.StringVar(&user, "user", "", "GitHub user")
	flag.Parse()
	if len(user) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify a GitHub user")
		flag.Usage()
		os.Exit(1)
	}

	home := homeDir()
	fileName := path.Join(home, ".github_"+user)
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			f2, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f2.Close()
			auth(user, f2)
		} else {
			// some other error opening the file
			fmt.Println(err)
			return
		}
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

	f.Close()
}

func auth(user string, file *os.File) {

	client := &http.Client{}

	scopes := []string{"repo", "public_repo"}
	post, err := json.Marshal(&Authorization{Note: "gha-app", Scopes: scopes})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", "https://api.github.com/authorizations", bytes.NewReader(post))

	devNull, err := os.Open("/dev/null")
	if err != nil {
		panic(err)
	}
	defer devNull.Close()
	fmt.Fprint(os.Stderr, "Password: ")
	pwd, err := speakeasy.FAsk(devNull, "")
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(user, pwd)
	fmt.Fprint(os.Stderr, "\n")

	fmt.Fprint(os.Stderr, "2FA token (optional): ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Fprint(os.Stderr, "\n")
	token = token[:len(token)-1]
	if token != "" {
		req.Header.Add("X-GitHub-OTP", token)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(resp.Body)
	authResp := make(map[string]interface{})
	err = decoder.Decode(&authResp)
	if err != nil {
		panic(err)
	}
	_ = resp.Body.Close()

	if v, ok := authResp["token"]; ok {
		fmt.Printf("%s\n", v)
		io.WriteString(file, v.(string))
	} else {
		fmt.Printf("%+v\n", authResp)
	}
}
