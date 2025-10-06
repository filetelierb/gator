package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/filetelierb/gator/internal/config"
	"github.com/filetelierb/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)


type Config = config.Config
type User = database.User

type state struct{
	db *database.Queries
	conf *Config
}
type command struct {
	name string
	args []string
}

type commands struct{
	commands map[string]func(*state,command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *commands) run(s *state, cmd command) error{
	if command, ok := c.commands[cmd.name]; !ok {
		return fmt.Errorf("command not found")
	} else {
		err := command(s,cmd)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *commands) register(name string, f func(*state,command) error) {
	c.commands[name] = f
}

func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return fmt.Errorf("no arguments passed")
    }
    userName := cmd.args[0]
    fmt.Printf("%v\n", userName)
    existingUser, err := s.db.GetUser(context.Background(), sql.NullString{String: userName, Valid: true})
    if err != nil {
        return err
    } else if existingUser == (database.User{}) {
        return fmt.Errorf("user not found")
    }
    if err = s.conf.SetUser(cmd.args[0]); err != nil {
        return err
    }
    fmt.Printf("Username set to %s\n", cmd.args[0])

    return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0{
		return fmt.Errorf("no argumments passed")
	}
	newUserData := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: sql.NullString{String: cmd.args[0], Valid: true},
	}
	newUser, err := s.db.CreateUser(context.Background(),newUserData)
	if err != nil {
		return err
	}
	fmt.Printf("User %s created successfully\n",newUser.Name.String)
	err = handlerLogin(s,cmd)
	if err != nil{
		return err
	}
	return nil
}

func handlerReset(s *state, cmd command) error{
	if err := s.db.ClearUserTable(context.Background()); err != nil{
		return err
	}
	fmt.Print("User table has been clear")
	return nil
}

func handlerGetUsers(s *state,cmd command) error{
	users, err := s.db.GetUsers(context.Background());
	if err != nil {
		return nil
	}
	for _, user := range users{
		current := ""
		if user.Name.String == s.conf.CurrentUserName{
			current = " (current)"
		}
		fmt.Printf("* %s%s\n",user.Name.String,current)

	}
	return nil
}

func handlerAgg(s *state, cmd command) error{
	feed, err := fetchFeed(context.Background(),"https://www.wagslane.dev/index.xml")
	if err != nil{
		return err
	}
	fmt.Printf("%v",feed)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error){
	req,err := http.NewRequestWithContext(ctx,"GET",feedURL,nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent","gator")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil{
		return nil, err
	}
	var rssFeed RSSFeed
	if err = xml.Unmarshal(data,&rssFeed); err != nil{
		return nil, err
	}
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i,v := range rssFeed.Channel.Item {
		v.Title = html.UnescapeString(v.Title)
		v.Description = html.UnescapeString(v.Description)
		rssFeed.Channel.Item[i] = v
	}
	
	return &rssFeed, nil

	

}

func handlerAddFeed(s *state,cmd command,u User) error{
	args := cmd.args
	if len(args) < 1{
		return fmt.Errorf("name and url arguments are missing")
	} else if len(args) < 2{
		return fmt.Errorf("url argument is missing")
	}
	
	
	feedRecord := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: sql.NullString{
			String: args[0],
			Valid: true,
		},
		Url: sql.NullString{
			String: args[1],
			Valid: true,
		},
		UserID: uuid.NullUUID{
			UUID: u.ID,
			Valid: true,
		},
	}
	newFeed, err := s.db.CreateFeed(context.Background(),feedRecord)
	if err != nil{
		return err
	}
	fmt.Printf("%v",newFeed)
	newCmd := cmd
	newCmd.args = cmd.args[1:]

	err = handlerFollow(s,newCmd,u)
	if err != nil{
		return fmt.Errorf("error crating follow feeds record: %v",err)
	}
	return nil


	
}

func handlerGetFeeds(s *state, cmd command) error{
	db := s.db

	allFeeds, err := db.GetFeeds(context.Background())
	if err != nil{
		return err
	}
	for _, feed := range allFeeds{
		fmt.Printf("Feed Name: %s\n", feed.Name.String)
		fmt.Printf("Feed Url: %s\n", feed.Url.String)
		fmt.Printf("Feed Creator: %s\n", feed.UserName.String)
	}
	return nil
}
func handlerFollow(s *state, cmd command,u User) error{
	args := cmd.args
	if len(args) < 1{
		return fmt.Errorf("no url was received")
	}
	
	feedUrl := sql.NullString{
		String: args[0],
		Valid: true,
	}
	db := s.db
	feedToFollow, err := db.GetFeed(context.Background(),feedUrl)
	if err != nil && err != sql.ErrNoRows{
		return err
	}
	feedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: uuid.NullUUID{
			UUID: u.ID,
			Valid: true,
		},
		FeedID: uuid.NullUUID{
			UUID: feedToFollow.ID,
			Valid: true,
		},
	}
	newFeedFollow, err := db.CreateFeedFollow(context.Background(),feedFollowParams)
	if err != nil{
		return err
	}
	fmt.Printf("%v",newFeedFollow)
	return nil
	
}

func handlerFollowing(s *state, cmd command, u User) error{
	
	db := s.db
	getFollowsForUserParams := uuid.NullUUID{
		UUID: u.ID,
		Valid: true,
	}
	feedFollows, err:= db.GetFeedFollowsForUser(context.Background(),getFollowsForUserParams)
	if err != nil && err != sql.ErrNoRows{
		return err
	}
	fmt.Printf("%s is following:\n\n",u.Name.String)
	for _,follow := range feedFollows{
		fmt.Printf("- %s\n",follow.Name.String)
	}
	return nil

}

func middlewareLoggedIn(handler func(s *state, cmd command, u User) error) func(*state, command) error {
    return func(s *state, cmd command) error {
        currentUserName := sql.NullString{
            String: s.conf.CurrentUserName,
            Valid:  true,
        }
        currentUser, err := s.db.GetUser(context.Background(), currentUserName)
        if err != nil {
            return err
        }
        return handler(s, cmd, currentUser)
    }
}



func main(){
	

	conf, err := config.Read()
	if err != nil {
		fmt.Printf("Couldn't READ the file for the FIRST time: %v", err)
	}
	db, err := sql.Open("postgres", conf.DbUrl)
	if err != nil{
		fmt.Printf("Error connecting to db at %s", conf.DbUrl)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	initState := state{
		db: dbQueries,
		conf: &conf,
	}
	commands := commands{
		commands: map[string]func(*state,command) error{},
	}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset",handlerReset)
	commands.register("users",handlerGetUsers)
	commands.register("agg",handlerAgg)
	commands.register("addfeed",middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerGetFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following",middlewareLoggedIn(handlerFollowing))

	
	args := os.Args[1:]

	if len(args) == 0{
		fmt.Print("No arguments were passed\n")
		os.Exit(1)
	}

	

	cmd := command{
		name: args[0],
		args: args[1:],
	}

	err = commands.run(&initState,cmd)
	if err != nil{
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}