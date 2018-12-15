package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cl, err := pubsub.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating Topic")
	_, err = cl.CreateTopic(ctx, "email")
	if err != nil {
		fmt.Println(err) // if this program is running inside a docker image, it will panic, otherwise -- create topic succeeds.
	}

	const top = "email"
	// Create a topic to subscribe to.

	fmt.Println("Create a topic to subscribe to")
	topic := cl.Topic(top)
	ok, err := topic.Exists(ctx)
	if err != nil {
		fmt.Println(err)
	}

	if ok {
		fmt.Println("Exists")
	}

	go InitiatePubsubPull()

	http.HandleFunc("/send_message", EndPoint) // set router
	err = http.ListenAndServe(":9090", nil)    // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func InitiatePubsubPull() {
	fmt.Println(" Spinning up a Go routine for the pubsub pull ")
	ctx := context.Background()
	proj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if proj == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}
	client, err := pubsub.NewClient(ctx, proj)

	if err != nil {
		fmt.Println(err)
	}

	const t = "email"
	// Create a topic to subscribe to.

	fmt.Println("Creating Topic")
	topic := client.Topic(t)

	fmt.Println("Done Creating Topic")
	ok, err := topic.Exists(ctx)
	fmt.Println(ok)
	CreateSubscription(client, topic, ctx, "email_sub")
	PullMessages(client, "email_sub", topic)
	if err != nil {
		fmt.Println(err)
	}
	if ok {
		fmt.Println("Exists")
	}
}

//Create a subscription
func CreateSubscription(client *pubsub.Client, topic *pubsub.Topic, ctx context.Context, subName string) {
	sub, err := client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Created subscription: %v\n", sub)
}

func PullMessages(client *pubsub.Client, name string, topic *pubsub.Topic) error {
	fmt.Println("In Pull messages")
	ctx := context.Background()
	creatingSub := client.Subscription(name)
	cctx, _ := context.WithCancel(ctx)
	errRecv := creatingSub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Printf("Got message: %q\n", string(msg.Data))
		msg.Ack()
	})
	if errRecv != nil {
		fmt.Println(errRecv)
	}
	return nil
}

//Publish message to topic
func PushToTopic(message string) error {
	fmt.Println("Push to Topic")
	ctx := context.Background()
	proj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if proj == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}
	client, err := pubsub.NewClient(ctx, proj)

	const t = "email"

	topic := client.Topic(t)
	ok, err := topic.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		fmt.Println("Topic Exists")
	} else {
		fmt.Println("Topic Doesn't Exist")
		return nil
	}

	var results []*pubsub.PublishResult

	res := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(message),
	})
	results = append(results, res)

	fmt.Println(res)
	for _, r := range results {
		_, err := r.Get(ctx)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func EndPoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint gets called")
	msg := r.URL.Query().Get("msg")

	fmt.Printf("Message to be sent to the topic %s \n", msg)

	PushToTopic(msg)
}
