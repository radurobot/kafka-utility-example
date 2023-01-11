package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
)

var (
	KafkaBroker = os.Getenv("KAFKA_BROKERS") // get brokers from environment variable

	RequestBody struct {
		Topic   string `json:"topic"`
		Message string `json:"message"`
	}
)

// Check if a kafka topic exists
func DoesTopicExist(topic string) bool {
	// create a new kafka admin client
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": KafkaBroker})
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

// Produce a message to a kafka topic
func ProduceMessage(topic string, message string) error {
	// check if topic exists
	if DoesTopicExist(topic) == false {
		return fmt.Errorf("topic does not exist") // return error if topic does not exist
	}
	// create a new producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": KafkaBroker}) // create a new producer
	if err != nil {
		return fmt.Errorf("failed to create producer: %v", err)
	}
	defer p.Close()
	// flush messages
	p.Flush(200)
	// produce message
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, // topic and partition
		Value:          []byte(message),                                                    // message
	}, nil) // callback function
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}
	return nil
}

// encrypt string to base64 crypto using AES
func encryptString(text string) string {
	var SecretKey = []byte(os.Getenv("SECRET_KEY")) // get secret key from environment variable
	plaintext := []byte(text)
	block, err := aes.NewCipher(SecretKey)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Create a Kafka topic
func CreateTopic(topic string) error {
	// check if topic exists
	if DoesTopicExist(topic) == true {
		// return error if topic already exists
		return fmt.Errorf("topic already exists")
	}
	// create a new kafka admin client
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": KafkaBroker})
	if err != nil {
		return fmt.Errorf("failed to create Admin client: %v", err)
	}
	defer adminClient.Close()
	// create a new topic
	result, err := adminClient.CreateTopics(
		context.Background(),
		// Provide topic specification
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
		// Admin options
		kafka.SetAdminOperationTimeout(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to create topic: %v", err)
	}
	for _, result := range result {
		fmt.Printf("TOPIC: %s\n", result)
	}
	return nil
}

func main() {
	router := gin.Default()
	// HTTP POST request to send message to kafka topic
	// Requires JSON BODY: {"topic": "topic_name", "message": "message"}
	router.POST("/send", func(c *gin.Context) {
		// extract requestBody from body
		if err := c.ShouldBindJSON(&RequestBody); err != nil {
			c.JSON(400, gin.H{"error": err.Error()}) // return error if json is not valid
			return
		}
		err := ProduceMessage(RequestBody.Topic, encryptString(RequestBody.Message)) // call producer function
		if err != nil {
			log.Fatalf("Failed to produce message: %v", err) // log error
		} else {
			log.Println("Message sent to kafka") // log message
		} // wait for confirmation from producer function
		c.JSON(200, gin.H{"message": "message sent to kafka"}) // return success message
	})
	// HTTP POST request to create a kafka topic
	// Requires JSON BODY: {"topic": "topic_name"}
	router.POST("/create", func(c *gin.Context) {
		// extract requestBody from body
		if err := c.ShouldBindJSON(&RequestBody); err != nil {
			// return error if json is not valid
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		// Create topic
		err := CreateTopic(RequestBody.Topic)
		if err != nil {
			log.Fatalf("Failed to create topic: %v", err) // log error
		} else {
			log.Println("Topic created") // log message
		} // wait for confirmation from producer function
		c.JSON(200, gin.H{"message": "topic created"}) // return success message
	})
	router.Run(":8081") // run router on port 8080
}
