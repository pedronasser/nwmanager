package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"nwmanager/database"
	"nwmanager/extractor"
	. "nwmanager/helpers"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.NewMongoDB(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}
	fmt.Println(db)

	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	f, err := os.Open("image.png")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	imageFile, err := uploadImage(ctx, client, f)
	if err != nil {
		log.Fatalf("failed to upload image: %v", err)
	}

	defer client.DeleteFile(ctx, imageFile.Name)

	memCol := db.Collection("members")

	members, err := extractor.GetMembersFromGuildImage(ctx, imageFile.URI)
	if err != nil {
		log.Fatalf("failed to extract members: %v", err)
	}

	for _, member := range members {
		fmt.Printf("Name: %s\nRank: %s\nReputation: %d\nLast Active: %s\n", member.Name, member.Rank, member.Reputation, member.LastActive)
		opts := &options.UpdateOptions{
			Upsert: Some(true),
		}
		_, err = memCol.UpdateOne(ctx, bson.M{
			"name": member.Name,
		}, bson.M{
			"$set": member,
		}, opts)
		if err != nil {
			log.Fatalf("failed to update members: %v", err)
		}
	}
}

func uploadImage(ctx context.Context, client *genai.Client, r io.Reader) (*genai.File, error) {
	opts := &genai.UploadFileOptions{MIMEType: "image/png"}
	file, err := client.UploadFile(ctx, "", r, opts)
	if err != nil {
		return nil, err
	}

	return file, nil
}
