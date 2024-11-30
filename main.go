package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	err := godotenv.Load("./env/.env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	var accessToken = os.Getenv("DISCORD_TOKEN")
	if accessToken == "" {
		log.Println("DISCORD_TOKEN is not set")
		return
	}
	discord, err := discordgo.New("Bot " + accessToken)

	if err != nil {
		log.Println("ERROR: failed to start session.\n", err.Error())
		return
	}

	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
	}
	defer discord.Close()

	log.Println("Start granma")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	clientId := os.Getenv("CLIENT_ID")
	u := m.Author
	ctx := context.Background()

	if len(m.Mentions) < 1 {
		return
	}

	if u.ID != clientId && m.Mentions[0].ID == clientId {
		resp := genMessage(m.Content, ctx)
		_, err := s.ChannelMessageSend(m.ChannelID, resp)
		if err != nil {

		}
	}
}

func genMessage(query string, ctx context.Context) string {
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_APIKEY")))
	if err != nil {
		return fmt.Sprint("Failed to cleate client: ", err.Error())
	}

	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are a granma."))
	prompt := genai.Text(query)

	resp, err := model.GenerateContent(ctx, prompt)
	if err != nil {
		return fmt.Sprint("ばあちゃんはメッセージの生成に失敗したよ！あんた変な言葉送ったんじゃないの？\n", err.Error())
	}

	return sprintResp(resp)

}

func sprintResp(resp *genai.GenerateContentResponse) string {
	var s string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			s += fmt.Sprintln(p)
		}
	}
	return s
}
