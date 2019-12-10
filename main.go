package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var serverPort = ":" + os.Getenv("PORT")
var client *mongo.Client

func main() {

	ctx, _ := context.WithTimeout(context.Background(), 240*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://derek:ajninsword@cluster0-swqnm.mongodb.net/test?retryWrites=true&w=majority")
	client, _ = mongo.Connect(ctx, clientOptions)

	defer client.Disconnect(ctx)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./client/build/static")))
	http.Handle("/static", staticHandler)

	http.HandleFunc("/api/add", addCard)
	http.HandleFunc("/api/decks", getAllDecks)
	http.HandleFunc("/api/makeDeck", addDeck)
	http.HandleFunc("/api/deleteDeck", deleteDeck)
	http.HandleFunc("/api/getCards", getCards)

	log.Println("Listening...")
	if serverPort == ":" {
		serverPort = ":8080"
	}
	log.Fatal(http.ListenAndServe(serverPort, nil))

}

type Test struct {
	Name string
}

type Deck struct {
	DeckName string
}

type Card struct {
	CardFront string
	CardBack  string
}

func addCard(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	//log.Println(string(body))
	//fmt.Fprintf(w, "%v", string(body))

	var t Test
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}
	//log.Println(t.Name)
}

func getAllDecks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "%v", databases)
}

func addDeck(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var deck Deck
	err = json.Unmarshal(body, &deck)
	fmt.Printf("%v", deck)

	newDatabase := client.Database(deck.DeckName)
	newCollection := newDatabase.Collection(deck.DeckName)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	newResult, err := newCollection.InsertOne(ctx, bson.D{
		{"cardFront", "Test card!"},
		{"cardBack", "Test card answer"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("inserted %v", newResult)
}

func deleteDeck(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var deck Deck
	err = json.Unmarshal(body, &deck)
	fmt.Printf("%v", deck)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	toBeDeletedDatabase := client.Database(deck.DeckName)
	toBeDeletedDatabase.Drop(ctx)

}

func getCards(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var deck Deck
	err = json.Unmarshal(body, &deck)
	fmt.Printf("%v", deck)

	database := client.Database(deck.DeckName)
	collection := database.Collection(deck.DeckName)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	// cursor, err := collection.FindOne(ctx, bson.M{})
	// if err != nil {
	// 	log.Fatal(err, "yo")
	// }
	var card bson.M
	var s Card
	if err = collection.FindOne(ctx, bson.M{}).Decode(&card); err != nil {
		log.Fatal(err)
	}

	bsonBytes, _ := bson.Marshal(card)
	bson.Unmarshal(bsonBytes, &s)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var cardMap []Card
	var cardTwo Card
	for cursor.Next(ctx) {
		var episode bson.M
		if err = cursor.Decode(&episode); err != nil {
			log.Fatal(err)
		}
		bsonBytes, _ := bson.Marshal(episode)
		bson.Unmarshal(bsonBytes, &cardTwo)
		cardMap = append(cardMap, cardTwo)
		fmt.Println(cardMap)
	}

	fmt.Fprintf(w, "%v", cardMap)
	// var cards []bson.M
	// if err = cursor.All(ctx, &cards); err != nil {
	// 	log.Fatal(err, "yoo")
	// }

}
