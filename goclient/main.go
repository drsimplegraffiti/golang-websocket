// package main
//
// import (
// 	"bufio"
// 	"context"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
//
// 	"github.com/coder/websocket"
// 	"github.com/coder/websocket/wsjson"
// )
//
// type Event struct {
// 	EventType string      `json:"event_type"`
// 	Payload   interface{} `json:"payload"`
// }
//
// func main() {
// 	url := flag.String("url", "ws://localhost:8082/api/ws", "WebSocket URL")
// 	token := flag.String("token", "", "JWT")
// 	name := flag.String("name", "User", "Display name")
//
// 	flag.Parse()
//
// 	if *token == "" {
// 		log.Fatal("token is required")
// 	}
//
// 	// Send Authorization header just like Postman.
// 	headers := http.Header{}
// 	headers.Set("Authorization", "Bearer "+*token)
//
// 	conn, resp, err := websocket.Dial(
// 		context.Background(),
// 		*url,
// 		&websocket.DialOptions{
// 			HTTPHeader: headers,
// 		},
// 	)
// 	if err != nil {
// 		if resp != nil {
// 			fmt.Println("Handshake failed:", resp.Status)
//
// 			if resp.Body != nil {
// 				body, _ := io.ReadAll(resp.Body)
// 				fmt.Println(string(body))
// 			}
// 		}
//
// 		log.Fatal(err)
// 	}
//
// 	defer conn.Close(websocket.StatusNormalClosure, "")
//
// 	fmt.Printf("[%s] Connected!\n\n", *name)
// 	fmt.Println("Paste a JSON event and press Enter.")
//
// 	// Read incoming events.
// 	go func() {
// 		for {
// 			var event Event
//
// 			err := wsjson.Read(context.Background(), conn, &event)
// 			if err != nil {
// 				log.Println("Disconnected:", err)
// 				os.Exit(0)
// 			}
//
// 			b, _ := json.MarshalIndent(event, "", "  ")
//
// 			fmt.Println("\n========== RECEIVED ==========")
// 			fmt.Println(string(b))
// 			fmt.Println("==============================")
// 		}
// 	}()
//
// 	scanner := bufio.NewScanner(os.Stdin)
//
// 	for scanner.Scan() {
// 		var event Event
//
// 		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
// 			fmt.Println("Invalid JSON")
// 			continue
// 		}
//
// 		if err := wsjson.Write(context.Background(), conn, event); err != nil {
// 			log.Println(err)
// 			return
// 		}
// 	}
// }

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Event struct {
	EventType string      `json:"event_type"`
	Payload   interface{} `json:"payload"`
}

func main() {
	url := flag.String("url", "ws://localhost:8082/api/ws", "WebSocket URL")
	token := flag.String("token", "", "JWT")
	name := flag.String("name", "User", "Display name")

	flag.Parse()

	if *token == "" {
		log.Fatal("token is required")
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+*token)

	conn, resp, err := websocket.Dial(
		context.Background(),
		*url,
		&websocket.DialOptions{
			HTTPHeader: headers,
		},
	)
	if err != nil {
		if resp != nil {
			fmt.Println("Handshake failed:", resp.Status)
			if resp.Body != nil {
				body, _ := io.ReadAll(resp.Body)
				fmt.Println(string(body))
			}
		}
		log.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	fmt.Printf("[%s] Connected!\n\n", *name)
	fmt.Println("Paste your JSON.")
	fmt.Println("Type END on a new line to send.")
	fmt.Println()

	go func() {
		for {
			var event Event

			if err := wsjson.Read(context.Background(), conn, &event); err != nil {
				log.Println("Disconnected:", err)
				os.Exit(0)
			}

			b, _ := json.MarshalIndent(event, "", "  ")

			fmt.Println("\n========== RECEIVED ==========")
			fmt.Println(string(b))
			fmt.Println("==============================")
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		var lines []string

		for scanner.Scan() {
			line := scanner.Text()

			if strings.TrimSpace(line) == "END" {
				break
			}

			lines = append(lines, line)
		}

		if len(lines) == 0 {
			continue
		}

		input := strings.Join(lines, "\n")

		var event Event
		if err := json.Unmarshal([]byte(input), &event); err != nil {
			fmt.Println("Invalid JSON:", err)
			continue
		}

		if err := wsjson.Write(context.Background(), conn, event); err != nil {
			log.Println(err)
			return
		}

		fmt.Println("✓ Sent")
	}
}

// package main
//
// import (
// 	"bufio"
// 	"context"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"strings"
//
// 	"github.com/coder/websocket"
// 	"github.com/coder/websocket/wsjson"
// )
//
// type Event struct {
// 	EventType string      `json:"event_type"`
// 	Payload   interface{} `json:"payload"`
// }
//
// func main() {
// 	url := flag.String("url", "ws://localhost:8082/api/ws", "WebSocket URL")
// 	token := flag.String("token", "", "JWT Token")
// 	name := flag.String("name", "User", "Display name")
//
// 	flag.Parse() // this line parses the command-line flags and assigns their
// 	// values to the corresponding variables.
//
// 	if *token == "" {
// 		log.Fatal("token is required")
// 	}
//
// 	headers := http.Header{}
// 	headers.Set("Authorization", "Bearer "+*token)
//
// 	conn, resp, err := websocket.Dial(
// 		context.Background(),
// 		*url,
// 		&websocket.DialOptions{
// 			HTTPHeader: headers,
// 		},
// 	)
// 	if err != nil {
// 		if resp != nil {
// 			fmt.Println("Handshake failed:", resp.Status)
//
// 			if resp.Body != nil {
// 				body, _ := io.ReadAll(resp.Body)
// 				fmt.Println(string(body))
// 			}
// 		}
//
// 		log.Fatal(err)
// 	}
//
// 	defer conn.Close(websocket.StatusNormalClosure, "")
//
// 	fmt.Printf("[%s] Connected!\n\n", *name)
//
// 	go func() {
// 		for {
// 			var event Event
//
// 			err := wsjson.Read(context.Background(), conn, &event)
// 			if err != nil {
// 				log.Println("Disconnected:", err)
// 				os.Exit(0)
// 			}
//
// 			b, _ := json.MarshalIndent(event, "", "  ")
//
// 			fmt.Println()
// 			fmt.Println("========== RECEIVED ==========")
// 			fmt.Println(string(b))
// 			fmt.Println("==============================")
// 			fmt.Print("> ")
// 		}
// 	}()
//
// 	printHelp()
//
// 	scanner := bufio.NewScanner(os.Stdin)
//
// 	for {
// 		fmt.Print("> ")
//
// 		if !scanner.Scan() {
// 			return
// 		}
//
// 		line := strings.TrimSpace(scanner.Text())
//
// 		if line == "" {
// 			continue
// 		}
//
// 		switch {
//
// 		case line == "/help":
// 			printHelp()
//
// 		case line == "/quit":
// 			fmt.Println("Bye!")
// 			return
//
// 		case line == "/json":
//
// 			fmt.Println("Paste JSON below.")
// 			fmt.Println("Finish by entering an empty line.")
//
// 			var lines []string
//
// 			for scanner.Scan() {
//
// 				t := scanner.Text()
//
// 				if strings.TrimSpace(t) == "" {
// 					break
// 				}
//
// 				lines = append(lines, t)
// 			}
//
// 			var event Event
//
// 			err := json.Unmarshal([]byte(strings.Join(lines, "\n")), &event)
// 			if err != nil {
// 				fmt.Println("Invalid JSON:", err)
// 				continue
// 			}
//
// 			err = wsjson.Write(context.Background(), conn, event)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
//
// 		case strings.HasPrefix(line, "/msg "):
//
// 			parts := strings.SplitN(line, " ", 4)
//
// 			if len(parts) != 4 {
// 				fmt.Println("Usage: /msg <receiverId> <privateId> <message>")
// 				continue
// 			}
//
// 			receiverID, err := strconv.ParseInt(parts[1], 10, 64)
// 			if err != nil {
// 				fmt.Println("Invalid receiver id")
// 				continue
// 			}
//
// 			privateID, err := strconv.ParseInt(parts[2], 10, 64)
// 			if err != nil {
// 				fmt.Println("Invalid private id")
// 				continue
// 			}
//
// 			event := Event{
// 				EventType: "message",
// 				Payload: map[string]any{
// 					"receiver_id":  receiverID,
// 					"private_id":   privateID,
// 					"message_type": "text",
// 					"content":      parts[3],
// 				},
// 			}
//
// 			if err := wsjson.Write(context.Background(), conn, event); err != nil {
// 				fmt.Println(err)
// 			}
//
// 		case strings.HasPrefix(line, "/typing "):
//
// 			parts := strings.Fields(line)
//
// 			if len(parts) != 4 {
// 				fmt.Println("Usage: /typing <receiverId> <privateId> on|off")
// 				continue
// 			}
//
// 			receiverID, err := strconv.ParseInt(parts[1], 10, 64)
// 			if err != nil {
// 				fmt.Println("Invalid receiver id")
// 				continue
// 			}
//
// 			privateID, err := strconv.ParseInt(parts[2], 10, 64)
// 			if err != nil {
// 				fmt.Println("Invalid private id")
// 				continue
// 			}
//
// 			event := Event{
// 				EventType: "typing",
// 				Payload: map[string]any{
// 					"receiver_id": receiverID,
// 					"private_id":  privateID,
// 					"is_typing":   strings.ToLower(parts[3]) == "on",
// 				},
// 			}
//
// 			if err := wsjson.Write(context.Background(), conn, event); err != nil {
// 				fmt.Println(err)
// 			}
//
// 		case strings.HasPrefix(line, "/read "):
//
// 			parts := strings.Fields(line)
//
// 			if len(parts) != 2 {
// 				fmt.Println("Usage: /read <messageId>")
// 				continue
// 			}
//
// 			id, err := strconv.ParseInt(parts[1], 10, 64)
// 			if err != nil {
// 				fmt.Println("Invalid message id")
// 				continue
// 			}
//
// 			event := Event{
// 				EventType: "read",
// 				Payload: map[string]any{
// 					"message_id": id,
// 				},
// 			}
//
// 			if err := wsjson.Write(context.Background(), conn, event); err != nil {
// 				fmt.Println(err)
// 			}
//
// 		case strings.HasPrefix(line, "/delivered "):
//
// 			parts := strings.Fields(line)
//
// 			if len(parts) != 2 {
// 				fmt.Println("Usage: /delivered <messageId>")
// 				continue
// 			}
//
// 			id, err := strconv.ParseInt(parts[1], 10, 64)
// 			if err != nil {
// 				fmt.Println("Invalid message id")
// 				continue
// 			}
//
// 			event := Event{
// 				EventType: "delivered",
// 				Payload: map[string]any{
// 					"message_id": id,
// 				},
// 			}
//
// 			if err := wsjson.Write(context.Background(), conn, event); err != nil {
// 				fmt.Println(err)
// 			}
//
// 		default:
// 			fmt.Println("Unknown command. Type /help")
// 		}
// 	}
// }
//
// func printHelp() {
// 	fmt.Println("Commands")
// 	fmt.Println("------------------------------------------------------")
// 	fmt.Println("/msg <receiverId> <privateId> <message>")
// 	fmt.Println("    Example:")
// 	fmt.Println("    /msg 2 1 Hello Bob!")
// 	fmt.Println()
//
// 	fmt.Println("/typing <receiverId> <privateId> on|off")
// 	fmt.Println("    Example:")
// 	fmt.Println("    /typing 2 1 on")
// 	fmt.Println()
//
// 	fmt.Println("/read <messageId>")
// 	fmt.Println("    Example:")
// 	fmt.Println("    /read 15")
// 	fmt.Println()
//
// 	fmt.Println("/delivered <messageId>")
// 	fmt.Println("    Example:")
// 	fmt.Println("    /delivered 15")
// 	fmt.Println()
//
// 	fmt.Println("/json")
// 	fmt.Println("    Enter raw JSON mode.")
// 	fmt.Println()
//
// 	fmt.Println("/help")
// 	fmt.Println("/quit")
// 	fmt.Println("------------------------------------------------------")
// }
