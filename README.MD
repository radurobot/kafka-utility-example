## Kafka utility

###### This repository shows an implementation of a kafka cluster and two utility REST APIs used to comunicate with the cluster. 

---
##### Setup:
Run script.sh with KAFKA_TOPICS and KAFKA_BROKER as environment variables. For setting up multiple topics delimiter them with a comma. 
To run the consumer & producer you can either build the docker file or use the shell script. Also SECRET_KEY for the encryption/decryption.

---- 

**Kafka Producer** has two functionalities:
* Create a kafka topic
* Send message to a topic

The **Kafka Consumer** uses concurrency to subscribe and listen to all topics and communicates with the REST API via a go channel. 

---
**All messages that go through the kafka cluster once are produced are encrypted and decrypted once they are consumed.**