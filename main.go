package main

import (
	"bufio"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dolmen-go/kittyimg"
	"github.com/nicklaw5/helix/v2"
)

func errhandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkTwitchCLI() {
	fmt.Println("Checking for Twitch-cli")
	_, err := exec.LookPath("twitch")
	if err != nil {
		fmt.Printf("didn't find 'twitch' or 'twitch-cli' executable\n")
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

type Stream struct {
	Title        string
	UserLogin    string
	UserName     string
	GameName     string
	GameID       string
	StartedAt    time.Time
	IsMature     bool
	ID           string
	ThumbnailURL string
	TagIDs       []string
	Type         string
	UserID       string
	ViewerCount  int
}

func newStream(stream helix.Stream) *Stream {
	var s Stream
	s.Title = stream.Title
	s.UserLogin = stream.UserLogin
	s.UserName = stream.UserName
	s.GameName = stream.GameName
	s.GameID = stream.GameID
	s.StartedAt = stream.StartedAt
	s.IsMature = stream.IsMature
	s.ID = stream.ID
	s.ThumbnailURL = strings.Replace(stream.ThumbnailURL, "{width}x{height}", "260x200", 1)
	s.TagIDs = stream.TagIDs
	s.Type = stream.Type
	s.UserID = stream.UserID
	s.ViewerCount = stream.ViewerCount
	return &s
}

func playStream(stream *Stream, quality string) {
	fmt.Println("Starting " + stream.UserName + "'s stream")
	cmd := exec.Command("streamlink", "-v", "--twitch-disable-hosting", "--title", "{author} - {title}", "https://www.twitch.tv/"+stream.UserName, quality)
	err := cmd.Run()
	errhandle(err)
}

func grabThumbnail(st *Stream) (filePath string) {
	filePath = st.UserName + "-thumb.png"
	cmd := exec.Command("curl", "-o", filePath, st.ThumbnailURL)
	cmd.Start()
	cmd.Wait()
	return filePath
}

func displayThumbnail(st *Stream) {
	imagePath := grabThumbnail(st)
	file, err := os.Open(imagePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	thumbNail, _, err := image.Decode(file)

	kittyimg.Fprint(os.Stdout, thumbNail)
}

func main() {

	checkTwitchCLI()
	client := createClient()

	apiResponse, err := client.GetFollowedStream(&helix.FollowedStreamsParams{
		// UserID: "spiritmancy",
		UserID: "80449608",
	})
	errhandle(err)

	// possible to inline streams var?
	streams := apiResponse.Data.Streams
	for index := range apiResponse.Data.Streams {
		st := newStream(streams[index])
		fmt.Println("printing struct")
		fmt.Printf("%+v\n", st)
		displayThumbnail(st)

	}

	// playStream(*streams[0], "best")
}

// // --can-handle-url URL
// // --config FILENAME
// // -j, --json
// //   Output JSON representations instead of the normal text output.
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
// TODOS:
// * obtain userid
// * increase player options (choose player, set retry settings, read streamlink configs), flow options for 'self configuration with out twitch-cli'
