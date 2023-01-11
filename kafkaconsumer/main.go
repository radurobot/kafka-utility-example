package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
)

var brokers = os.Getenv("KAFKA_BROKERS") // get brokers from environment variable

// Check if a kafka topic exists
func DoesTopicExist(topic string) bool {
	// create a new kafka admin client
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		log.Fatalf("Failed to create Admin client: %v", err)
	}
	defer adminClient.Close()
	// return all topics
	topics, err := adminClient.GetMetadata(nil, true, 5000)
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}
	_, ok := topics.Topics[topic] // check if topic exists
	return ok

}

// Subscribe & consume messages from a topic that passes the message through a channel for REST API to consume
func KafkaConsumer(topic string, mesageChannel chan string) (string, error) {
	// check if topic exists
	topicExists := DoesTopicExist(topic) // call function to check if topic exists
	if topicExists == false {
		return "", fmt.Errorf("topic does not exist") // return error if topic does not exist
	}
	// create a new consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{ // create a new consumer
		"bootstrap.servers": brokers,    // brokers
		"group.id":          "myGroup",  // group id
		"auto.offset.reset": "earliest", // offset
	})
	if err != nil {
		return "", fmt.Errorf("failed to create consumer: %v", err)
	}
	defer c.Close() // close consumer when done
	// subscribe to topic
	c.SubscribeTopics([]string{topic}, nil) // subscribe to topic
	// use mesageChannel to send the message to the REST API
	for {
		fmt.Println("Waiting for message...") // print message
		msg, err := c.ReadMessage(-1)         // read message
		if err == nil {
			sendingdata := decryptString(string(msg.Value)) // decrypt message
			// send message to mesageChannel
			mesageChannel <- sendingdata

		} else {
			fmt.Println("Error reading message from Kafka: ", err) // print error
		}
	}
}

// decrypt from base64 to decrypted string
func decryptString(cryptoText string) string {
	var key = []byte(os.Getenv("SECRET_KEY")) // get secret key from environment variable
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("Error creating cipher block: ", err) // print error
	}
	if len(ciphertext) < aes.BlockSize {
		log.Println("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}

// Get all topics from kafka cluster
func GetAllTopics() []string {
	// create a new kafka admin client
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		log.Fatalf("Failed to create Admin client: %v", err)
	}
	defer adminClient.Close()
	// return all topics
	topics, err := adminClient.GetMetadata(nil, true, 5000)
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}
	var topicList []string
	for topic := range topics.Topics {
		topicList = append(topicList, topic)
	}
	return topicList
}

func main() {
	// get all topics from kafka broker
	topics := GetAllTopics()
	log.Println("Topics: ", topics)
	// create a map of channels
	var channelList map[string]chan string
	// initialize map of channels
	channelList = make(map[string]chan string)
	// loop through all topics
	for _, topic := range topics {
		if topic != "__consumer_offsets" { // ignore the __consumer_offsets topic
			// create a channel for each topic
			channelList[topic] = make(chan string)
			// call function to concurrently subscribe & listen to messages from each topic
			go KafkaConsumer(topic, channelList[topic])
		}
	}

	// create a REST API to read the messages from the channels
	router := gin.Default()
	// endpoint to read messages based on topic
	router.GET("/messages/:topic", func(c *gin.Context) {
		topicquery := c.Param("topic")
		fmt.Println("Topic: ", topicquery)
		// check if topic exists
		topicExists := DoesTopicExist(topicquery)
		if topicExists == false {
			c.JSON(http.StatusNotFound, gin.H{"error": "topic does not exist"}) // return error if topic does not exist
			return
		} else {
			// check if topic has a channel
			if _, ok := channelList[topicquery]; ok {
				// read message from channel
				select {
				case msg := <-channelList[topicquery]: // read message from channel
					c.JSON(http.StatusOK, gin.H{"message": msg}) // return message
				default:
					c.JSON(http.StatusNotFound, gin.H{"error": "no message received"}) // return error if no message received
				}
			} else { // if topic does not have a channel
				c.JSON(http.StatusNotFound, gin.H{"error": "no channel for topic"})
			}

		}
	})
	// Start router
	router.Run(":8080")
}
