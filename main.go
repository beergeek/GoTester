package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type PayloadType string

const (
	NoPayload          PayloadType = "No Payload"
	GenericPayload     PayloadType = "Generic Payload"
	PotentialMalicious PayloadType = "Potentially Malicious Payload"
)

type Conversation struct {
	Source      string
	Destination string
	Protocol    string
	PayloadType PayloadType
	HasPayload  bool
}

func analyzePayload(payload []byte) PayloadType {
	// This is a basic check. In real scenarios, more sophisticated analysis would be required.
	payloadStr := string(payload)

	// Check for common patterns of malicious content
	maliciousPatterns := []string{"malware", "virus", "exploit", "attack", "java", "kubernetes"}
	for _, pattern := range maliciousPatterns {
		if strings.Contains(strings.ToLower(payloadStr), pattern) {
			return PotentialMalicious
		}
	}

	if len(payload) > 0 {
		return GenericPayload
	}

	return NoPayload
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <pcap file>\n", os.Args[0])
	}
	filename := os.Args[1]

	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	conversations := make(map[string]*Conversation)

	for packet := range packetSource.Packets() {
		networkLayer := packet.NetworkLayer()
		if networkLayer == nil {
			continue
		}

		transportLayer := packet.TransportLayer()
		if transportLayer == nil {
			continue
		}

		source := networkLayer.NetworkFlow().Src().String()
		destination := networkLayer.NetworkFlow().Dst().String()
		protocol := transportLayer.LayerType().String()
		payload := transportLayer.LayerPayload()
		payloadType := analyzePayload(payload)

		convKey := fmt.Sprintf("%s:%s:%s", source, destination, protocol)
		if conv, exists := conversations[convKey]; exists {
			if payloadType == PotentialMalicious || (payloadType == GenericPayload && conv.PayloadType != PotentialMalicious) {
				conv.PayloadType = payloadType
			}
			if len(payload) > 0 {
				conv.HasPayload = true
			}
		} else {
			conversations[convKey] = &Conversation{
				Source:      source,
				Destination: destination,
				Protocol:    protocol,
				PayloadType: payloadType,
				HasPayload:  len(payload) > 0,
			}
		}
	}

	var output []Conversation
	for _, conv := range conversations {
		output = append(output, *conv)
	}

	fmt.Println("Conversations:")
	for _, conv := range output {
		fmt.Printf("Source: %s, Destination: %s, Protocol: %s, Has Payload: %v, Payload Type: %s\n",
			conv.Source, conv.Destination, conv.Protocol, conv.HasPayload, conv.PayloadType)
	}
}
