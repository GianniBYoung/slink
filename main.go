package main

// /home/gianni/.config/twitch-cli/.twitch-cli.env

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/nicklaw5/helix/v2"
)

func errhandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getCredentials() (cid, secret, access, refresh string) {
	scanner := bufio.NewScanner(os.Stdin)

	home := os.Getenv("HOME")
	twitchConfig := (home + "/.config/twitch-cli/.twitch-cli.env")
	if _, err := os.Stat(twitchConfig); err == nil {
		fmt.Println("Twitch credentials found")
	} else {

		fmt.Println("Twitch credentials not found, please Enter ClientID")
		for scanner.Scan() {
			cid = scanner.Text()
			break
		}

		fmt.Println("Please Enter ClientSecret")
		for scanner.Scan() {
			secret = scanner.Text()
			break
		}

		// consider cleaning this up by reducing line
		cmd := exec.Command("twitch", "configure", "-i", cid, "-s", secret)
		err := cmd.Start()
		err = cmd.Wait()
		errhandle(err)
		cmd = exec.Command("twitch", "token", "-u", "-s", "user:read:follows")
		err = cmd.Start()
		err = cmd.Wait()
		errhandle(err)
	}

	f, err := os.Open(twitchConfig)
	errhandle(err)
	defer f.Close()
	scanner = bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ACCESSTOKEN") {
			access = strings.Replace(line, "ACCESSTOKEN=", "", 1)
		} else if strings.Contains(line, "REFRESHTOKEN") {
			refresh = strings.Replace(line, "REFRESHTOKEN=", "", 1)
		} else if strings.Contains(line, "CLIENTID") {
			cid = strings.Replace(line, "CLIENTID=", "", 1)
		} else {
			secret = strings.Replace(line, "CLIENTSECRET=", "", 1)
		}
	}

	return cid, secret, access, refresh
}
func createClient() *helix.Client {
	clientID, clientSecret, accessToken, _ := getCredentials()

	client, err := helix.NewClient(&helix.Options{
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURI:     "http://localhost:3000",
		UserAccessToken: accessToken,
	})
	errhandle(err)
	fmt.Println("Validating Access Token")
	isValid, resp, err := client.ValidateToken(accessToken)
	if err != nil {
		fmt.Println(resp)
	} else if isValid {
		fmt.Println("Access Token is Valid! UwU")
	} else {
		fmt.Println(resp)
	}
	return client
}

func main() {
	// cmd := exec.Command("streamlink", "-v", "--title", "{author} - {title}", "https://www.twitch.tv/sodapoppin", "best")
	// // cmd := exec.Command("streamlink", "https://www.twitch.tv/sodapoppin", "best")
	// err := cmd.Run()
	// errhandle(err)
	checkTwitchCLI()
	client := createClient()

	apiResponse, err := client.GetFollowedStream(&helix.FollowedStreamsParams{
		UserID: "80449608",
	})
	errhandle(err)

	streams := apiResponse.Data.Streams
	fmt.Println(streams)

}

// }

// // streamlink -v --title "{author} - {title}" --twitch-disable-hosting https://www.twitch.tv/sodapoppin best

// // --can-handle-url URL
// // --config FILENAME
// // -j, --json
// //   Output JSON representations instead of the normal text output.
// // --version-check
// //   Runs a version check and exits.
// // -a ARGUMENTS, --player-args ARGUMENTS
// //   This option allows you to customize the default arguments which are put
// //   together with the value of --player to create a command to execute.

// // --player-continuous-http
// //   Make the player read the stream through HTTP, but unlike --player-http
// //   it will continuously try to open the stream if the player requests it.

// //   This makes it possible to handle stream disconnects if your player is
// //   capable of reconnecting to a HTTP stream. This is usually done by
// //   setting your player to a "repeat mode".

// // --default-stream STREAM -> aka best

// // --retry-max COUNT
// //   When using --retry-streams, stop retrying the fetch after COUNT retry
// //   attempt(s). Fetch will retry infinitely if COUNT is zero or unset.

// //   If --retry-max is set without setting --retry-streams, the delay between
// //   retries will default to 1 second.

// // --retry-open ATTEMPTS
// //   After a successful fetch, try ATTEMPTS time(s) to open the stream until
// //   giving up.

func checkTwitchCLI() {
	fmt.Println("Checking for Twitch-cli")
	_, err := exec.LookPath("twitch")
	if err != nil {
		fmt.Printf("didn't find 'twitch' or 'twitch-cli' executable\n")
	}
}
