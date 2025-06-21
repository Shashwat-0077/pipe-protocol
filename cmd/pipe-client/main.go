package main

import (
	"flag"
	"log"
	"pipe-protocol/pkg/client"
	"strings"
)

func main() {
	var (
		host       = flag.String("host", "localhost", "Server host")
		port       = flag.Int("port", 2525, "Server port")
		from       = flag.String("from", "", "From address")
		to         = flag.String("to", "", "To addresses (comma separated)")
		subject    = flag.String("subject", "", "Message subject")
		body       = flag.String("body", "", "Message body")
		serverName = flag.String("server", "client.local", "Client server name")
	)
	flag.Parse()

	if *from == "" || *to == "" || *subject == "" || *body == "" {
		log.Fatal("All fields (from, to, subject, body) are required")
	}

	toAddresses := strings.Split(*to, ",")
	for i, addr := range toAddresses {
		toAddresses[i] = strings.TrimSpace(addr)
	}

	client := client.NewClient(*serverName)

	headers := map[string]string{
		"message-id": "<test@client.local>",
		"date":       "2024-01-01T12:00:00Z",
	}

	err := client.SendMessage(*host, *port, *from, toAddresses, *subject, *body, headers)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	log.Println("Message sent successfully!")
}
