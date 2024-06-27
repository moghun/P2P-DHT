// P2PSEC - SS24 - Project Phase 1
// Muhammed Orhun Gale & Kerem Kurnaz
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
)

func prepareMessageBody(challenge []byte, teamNumber uint16, projectChoice uint16, nonce uint64, email, firstname, lastname, username string) []byte {
	var body bytes.Buffer
	binary.Write(&body, binary.BigEndian, challenge)
	binary.Write(&body, binary.BigEndian, teamNumber)
	binary.Write(&body, binary.BigEndian, projectChoice)
	binary.Write(&body, binary.BigEndian, nonce) // nonce is now part of the body for the hash computation
	body.WriteString(email + "\r\n")
	body.WriteString(firstname + "\r\n")
	body.WriteString(lastname + "\r\n")
	body.WriteString(username)
	return body.Bytes()
}

// createMessageHeader creates the header for the message with the given size and message type.
func createMessageHeader(messageSize uint16, messageType uint16) []byte {
	var messageHeader bytes.Buffer
	binary.Write(&messageHeader, binary.BigEndian, messageSize)
	binary.Write(&messageHeader, binary.BigEndian, messageType)
	return messageHeader.Bytes()
}

func main() {
    // Connect to the server
    conn, err := net.Dial("tcp", "p2psec.net.in.tum.de:13337")
    if err != nil {
        log.Println("Error connecting:", err)
        os.Exit(1)
    }
    defer conn.Close()

	// Read the challenge from the server.
	header := make([]byte, 4) // assuming the server sends a 4-byte header
	challenge := make([]byte, 8)
	reader := bufio.NewReader(conn)
	if _, err := reader.Read(header); err != nil {
		log.Printf("Error reading header: %v\n", err)
		os.Exit(1)
	}
	if _, err := reader.Read(challenge); err != nil {
		log.Printf("Error reading challenge: %v\n", err)
		os.Exit(1)
	}

	log.Print("ENROLL INIT SUCCESS", header, challenge, "\n")


	// Now, let's create the message to send for ENROLL REGISTER.

	// Assume we've found a valid nonce (just an example here).

	// MESSAGE SETTINGS
	//***********************************************************************************************
	//***********************************************************************************************
	//***********************************************************************************************

	const messageTypeEnrollRegister uint16 = 681 // The message type for "ENROLL INIT"
	email := "morhun.gale@tum.de"
	firstname := "Muhammed Orhun"
	lastname := "Gale"
	username := "moghun"
	teamNumber := uint16(0)
	projectChoice := uint16(4963) //DHT
	//************************************************************************************************
	//***********************************************************************************************
	//***********************************************************************************************


	// Find a valid nonce.
	nonce, err := findValidNonce(challenge, teamNumber, projectChoice, email, firstname, lastname, username)
	if err != nil {
		log.Printf("Error finding valid nonce: %v\n", err)
		os.Exit(1)
	}

	// Prepare the body of the message.
	body := prepareMessageBody(challenge, teamNumber, projectChoice, nonce, email, firstname, lastname, username)

	// Calculate message size (size of body plus 4 bytes for the header itself).
	messageSize := uint16(len(body) + 4) // The header contains 2 bytes for size and 2 bytes for message type.

	// Create the header with the size and message type.
	messageHeader := createMessageHeader(messageSize, messageTypeEnrollRegister)

	// Concatenate the header and body to form the full message.
	var fullMessage bytes.Buffer
	fullMessage.Write(messageHeader)
	fullMessage.Write(body)

	log.Print("Sending the message, size: ", messageSize)

	// Send the full message.
	if _, err := conn.Write(fullMessage.Bytes()); err != nil {
		log.Printf("Error sending ENROLL REGISTER message: %v\n", err)
		os.Exit(1)
	}

	responseHeader := make([]byte, 4)
	_, err = reader.Read(responseHeader)
	if err != nil {
		log.Printf("Error reading response header: %v\n", err)
		os.Exit(1)
	}

	// Determine the message type based on the responseHeader.
	messageType := binary.BigEndian.Uint16(responseHeader[2:4])

	switch messageType {
	case 682: // Assuming 682 corresponds to ENROLL SUCCESS
		teamNumber, err := handleEnrollSuccess(reader)
		if err != nil {
			log.Printf("Error handling ENROLL SUCCESS: %v\n", err)
			os.Exit(1)
		}
		log.Printf("Enrollment successful! Team number: %d\n", teamNumber)

	case 683: // Assuming 683 corresponds to ENROLL FAILURE
		errorNumber, errorDescription, err := handleEnrollFailure(reader)
		if err != nil {
			log.Printf("Error handling ENROLL FAILURE: %v\n", err)
			os.Exit(1)
		}
		log.Printf("Enrollment failed! Error number: %d, Description: %s\n", errorNumber, errorDescription)

	default:
		log.Printf("Received unknown message type: %d\n", messageType)
	}
}

func findValidNonce(challenge []byte, teamNumber uint16, projectChoice uint16, email, firstname, lastname, username string) (uint64, error) {
    var wg sync.WaitGroup
    nonceChan := make(chan uint64)
    errChan := make(chan error)
    doneChan := make(chan struct{})

    // Prepare the parts of the message that do not change
    initialPart := prepareInitialPart(challenge, teamNumber, projectChoice)
    userDetails := prepareUserDetails(email, firstname, lastname, username)
	log.Print("Correct SHA24 hash is getting searched")

    // Start multiple goroutines based on the number of available CPUs
    for i := 0; i < runtime.NumCPU(); i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            searchForNonce(initialPart, userDetails, nonceChan, errChan, doneChan)
        }()
    }

    // Close the nonce channel once all goroutines complete
    go func() {
        wg.Wait()
        close(nonceChan)
        close(errChan)
    }()

    // Return the first valid nonce or error found
    select {
    case nonce := <-nonceChan:
        close(doneChan) // Tell all goroutines to stop
        return nonce, nil
    case err := <-errChan:
        return 0, err
    }
}

func searchForNonce(initialPart, userDetails []byte, nonceChan chan uint64, errChan chan error, doneChan chan struct{}) {
    for {
        select {
        case <-doneChan:
            return // Exit if done
        default:
            // Generate a random nonce
            nonceBytes := make([]byte, 8)
            if _, err := rand.Read(nonceBytes); err != nil {
                errChan <- fmt.Errorf("error generating nonce: %v", err)
                return
            }

            // Combine parts to form the full message
            var data bytes.Buffer
            data.Write(initialPart)
            data.Write(nonceBytes)
            data.Write(userDetails)

            // Compute SHA256 hash
            hash := sha256.Sum256(data.Bytes())

            // Check if the first 30 bits of the hash are zero
            if hash[0] == 0 && hash[1] == 0 && hash[2] == 0 && (hash[3]&0xC0) == 0 {
                nonce := binary.BigEndian.Uint64(nonceBytes)
                nonceChan <- nonce
				log.Print("Correct SHA24 hash founded: ", hash, "\nNonce: ", nonce)
                return
            }
        }
    }
}

func prepareInitialPart(challenge []byte, teamNumber, projectChoice uint16) []byte {
    var buffer bytes.Buffer
    binary.Write(&buffer, binary.BigEndian, challenge)
    binary.Write(&buffer, binary.BigEndian, teamNumber)
    binary.Write(&buffer, binary.BigEndian, projectChoice)
    return buffer.Bytes()
}

func prepareUserDetails(email, firstname, lastname, username string) []byte {
    var buffer bytes.Buffer
    buffer.WriteString(email + "\r\n")
    buffer.WriteString(firstname + "\r\n")
    buffer.WriteString(lastname + "\r\n")
    buffer.WriteString(username)
    return buffer.Bytes()
}


// handleEnrollSuccess parses the ENROLL SUCCESS message and returns the team number.
func handleEnrollSuccess(reader *bufio.Reader) (uint16, error) {
	// Read the remaining part of the message.
	reserved := make([]byte, 2) // size includes the header; we already read 4 bytes of the header.
	_, err := reader.Read(reserved)
	if err != nil {
		return 0, fmt.Errorf("error reading ENROLL SUCCESS message: %v", err)
	}
	teamNumberBytes := make([]byte, 2) // size includes the header; we already read 4 bytes of the header.
	_, err = reader.Read(teamNumberBytes)
	if err != nil {
		return 0, fmt.Errorf("error reading ENROLL SUCCESS message: %v", err)
	}
	// Parse the team number from the message.
	// The team number is at index 2 of the message because the first two bytes are reserved.
	teamNumber := binary.BigEndian.Uint16(teamNumberBytes)
	return teamNumber, nil
}

// handleEnrollFailure parses the ENROLL FAILURE message and returns the error details.
func handleEnrollFailure(reader *bufio.Reader) (uint16, string, error) {
	header := make([]byte, 4) // Read the 4-byte header which contains the size
	_, err := reader.Read(header)
	if err != nil {
		return 0, "", fmt.Errorf("error reading ENROLL FAILURE header: %v", err)
	}

	// Parse the header to get the size of the message.
	size := binary.BigEndian.Uint16(header[0:2])

	// Read the remaining part of the message.
	message := make([]byte, size-4) // size includes the header; we already read 4 bytes of the header.
	_, err = reader.Read(message)
	if err != nil {
		return 0, "", fmt.Errorf("error reading ENROLL FAILURE message: %v", err)
	}

	// Parse the error number from the message.
	// The error number is at index 2 of the message because the first two bytes are reserved.
	errorNumber := binary.BigEndian.Uint16(message[2:4])

	// The error description is the remaining part of the message after the error number.
	errorDescription := string(message[4:])

	return errorNumber, errorDescription, nil
}
