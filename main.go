package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	config "mixin-bot/config"

	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	uuid "github.com/satori/go.uuid"
)

// CNBAssetID is the CNB's ID in Mixin Network
const CNBAssetID = "965e5c6e-434c-3fa9-b780-c50f43cd955c"

const defaultResponse = "Hello Welcome To Our Network!"

var client *bot.BlazeClient

// Handler is an implementation for interface bot.BlazeListener
// check out the url for more details: https://github.com/MixinNetwork/bot-api-go-client/blob/master/blaze.go#L89.
type Handler struct{}

// OnMessage is a general method of bot.BlazeListener
func (r Handler) OnMessage(ctx context.Context, msgView bot.MessageView, botID string) error {
	// I handle PLAIN_TEXT message only and make sure respond to current conversation.
	if msgView.Category == bot.MessageCategoryPlainText &&
		msgView.ConversationId == bot.UniqueConversationId(config.GetConfig().ClientID, msgView.UserId) {
		var data []byte
		var err error
		if data, err = base64.StdEncoding.DecodeString(msgView.Data); err != nil {
			log.Panicf("Error: %s\n", err)
			return err
		}
		inst := string(data)
		log.Printf("I got a message from %s, it said: `%s`\n", msgView.UserId, inst)

		if "Hi" == inst || "hi" == inst {
			// Sync? Ack!
			Respond(ctx, msgView, "Hello sir! how can i assist you.")
			Respond(ctx, msgView, "You can Ask me for 'Send money' and 'assets'.")
		} else if "Send money" == inst || "Send Money" == inst {
			// Show me the money. Okay.
			Transfer(ctx, msgView)
		} else if "assets" == inst || "Assets" == inst {
			// Show me your assets. Okay.
			ShowAssets(ctx, msgView)
		} else {
			Respond(ctx, msgView, defaultResponse)
		}
	}
	return nil
}

// Transfer 1.024 CNB to the user who having a conversation with bot.
func Transfer(ctx context.Context, msgView bot.MessageView) {
	payload := bot.TransferInput{
		AssetId:     CNBAssetID,
		RecipientId: msgView.UserId,
		Amount:      number.FromString("1"),
		TraceId:     uuid.Must(uuid.NewV4()).String(),
		Memo:        "Enjoy!",
	}
	err := bot.CreateTransfer(ctx, &payload,
		config.GetConfig().ClientID,
		config.GetConfig().SessionID,
		config.GetConfig().PrivateKey,
		config.GetConfig().Pin,
		config.GetConfig().PinToken,
	)
	if err != nil {
		Respond(ctx, msgView, fmt.Sprintf("Oops, %s\n", err))
	}
}

// ShowAssets show all my assets
func ShowAssets(ctx context.Context, msgView bot.MessageView) {
	var accessToken string
	var err error
	var assets []bot.Asset
	// generate an access token
	accessToken, err = bot.SignAuthenticationToken(
		config.GetConfig().ClientID,
		config.GetConfig().SessionID,
		config.GetConfig().PrivateKey, "GET", "/assets", "")
	if err != nil {
		Respond(ctx, msgView, fmt.Sprintf("Oops, %s\n", err))
		return
	}
	// get all assets.
	assets, err = bot.AssetList(ctx, accessToken)
	if err != nil {
		Respond(ctx, msgView, fmt.Sprintf("Oops, %s\n", err))
		return
	}
	var assetsBuffer bytes.Buffer
	for _, asset := range assets {
		assetsBuffer.WriteString(fmt.Sprintf("%s: %s\n", asset.Symbol, asset.Balance))
	}
	Respond(ctx, msgView, fmt.Sprintf("My Assets:\n%s", assetsBuffer.String()))
}

// Respond to user.
func Respond(ctx context.Context, msgView bot.MessageView, msg string) {
	if err := client.SendPlainText(ctx, msgView, msg); err != nil {
		log.Panicf("Error: %s\n", err)
	}
}

func main() {
	ctx := context.Background()
	log.Println("start bot")
	handler := Handler{}

	// Create a bot client
	client = bot.NewBlazeClient(config.GetConfig().ClientID, config.GetConfig().SessionID, config.GetConfig().PrivateKey)

	// Start the loop
	for {
		if err := client.Loop(ctx, handler); err != nil {
			log.Printf("Error: %v\n", err)
		}
		log.Println("connection loop end")
		time.Sleep(time.Second)
	}
}
