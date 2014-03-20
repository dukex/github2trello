package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type User struct {
	Name     string
	Email    string
	Username string
}

type Commit struct {
	Id         string
	Distinct   bool
	Message    string
	Timestamp  string
	Url        string
	Author     User
	Committer  User
	Added      []string
	Removed    []string
	Modified   []string
	Repository Repository
	Pusher     User
}

type Repository struct {
	Id          int
	Name        string
	Url         string
	Description string
	Homepage    string
	Watchers    int
	Stargazers  int
	Forks       int
	Fork        bool
	Size        int
	Owner       struct {
		Name  string
		Email string
	}
	Private       bool
	Open_issues   int
	Has_issues    bool
	Has_downloads bool
	Has_wiki      bool
	Language      string
	Created_at    int
	Pushed_at     int
	Master_branch string
	Organization  string
}

type PaylodParams struct {
	Ref     string
	After   string
	Before  string
	Created bool
	Delete  bool
	Forced  bool
	Compare string
	Commits []Commit
}

func PayloadHandler(w http.ResponseWriter, r *http.Request) {
	var params PaylodParams
	json.Unmarshal([]byte(r.FormValue("payload")), &params)

	if len(params.Commits) > 0 {
		go ToTrello(params)
	}

	fmt.Println("PARAMS:", params)
}

var (
	TRELLO_ENDPOINT = "https://api.trello.com/1"
	TRELLO_API_KEY  = os.Getenv("TRELLO_API_KEY")
	TRELLO_TOKEN    = os.Getenv("TRELLO_TOKEN")
	cardRegexp      = regexp.MustCompile(`#(\w+)`)
)

func ToTrello(params PaylodParams) {
	for _, commit := range params.Commits {
		cards := cardRegexp.FindAllStringSubmatch(commit.Message, -1)
		for _, card := range cards {
			fmt.Println(card)
			trelloCommentCommit(card[1], commit)
		}
	}
}

func trelloCommentCommit(cardId string, commit Commit) {
	text := commit.Committer.Name + " push the commit '" + commit.Message + "'[" + commit.Id + "](" + commit.Url + ")"
	data := url.Values{}
	data.Set("text", text)
	fmt.Println(text)
	sendPost("/cards/"+cardId+"/actions/comments", data)
}

func sendPost(resource string, data url.Values) {
	data.Set("key", TRELLO_API_KEY)
	data.Set("token", TRELLO_TOKEN)
	api_url := TRELLO_ENDPOINT + resource + "?" + data.Encode()

	fmt.Println("POST " + api_url)
	fmt.Println(http.PostForm(api_url, url.Values{}))
}

func main() {
	PORT := os.Getenv("PORT")
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/payload", PayloadHandler).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	h := handlers.LoggingHandler(os.Stdout, r)

	http.Handle("/", h)
	server := &http.Server{
		Addr:           ":" + PORT,
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Starting server on", PORT)
	log.Fatal(server.ListenAndServe())
}
