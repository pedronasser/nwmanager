package extractor

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"strconv"
	"strings"

	"nwmanager/types"
	. "nwmanager/types"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func promptFromImage(ctx context.Context, imageURI, prompt, modelName string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return "", fmt.Errorf("unable to create client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	model.SetTemperature(1)
	model.SetMaxOutputTokens(8192)
	model.SetTopP(0.95)
	// // configure the safety settings thresholds
	// model.SafetySettings = []*genai.SafetySetting{}

	// Given an image file URL, prepare image file as genai.Part
	img := genai.FileData{
		MIMEType: mime.TypeByExtension(imageURI),
		URI:      imageURI,
	}

	res, err := model.GenerateContent(ctx, img, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("unable to generate contents: %w", err)
	}

	return fmt.Sprintf("%s", res.Candidates[0].Content.Parts[0]), nil
}

func GetMembersFromGuildImage(ctx context.Context, imageURI string) ([]GuildMember, error) {
	resp, err := promptFromImage(ctx, imageURI,
		"List all character names, their ranks, reputation and current location. One character per line, and separate each field with ;.",
		"gemini-1.5-flash-001",
	)
	if err != nil {
		log.Fatalf("failed to generate multimodal content: %v", err)
	}

	var members []GuildMember

	scanner := bufio.NewScanner(strings.NewReader(resp))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ";")
		if len(parts) != 4 {
			continue
		}

		rep, err := strconv.ParseInt(strings.ReplaceAll(strings.TrimSpace(parts[2]), ",", ""), 10, 64)
		if err != nil {
			continue
		}

		members = append(members, GuildMember{
			Name:       strings.TrimSpace(parts[0]),
			Rank:       types.ParseRank(strings.TrimSpace(parts[1])),
			Reputation: int(rep),
			LastActive: types.ParseLastActive(strings.TrimSpace(parts[3])),
		})
	}

	return members, nil
}
