package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// will be set during build via ldflags
var DefaultClientID = "PLACEHOLDER_CLIENT_ID"

// Discord RPC message types
const (
	OpHandshake = iota
	OpFrame
)

// HandshakeMessage RPC structures
type HandshakeMessage struct {
	Version  string `json:"v"`
	ClientID string `json:"client_id"`
}

type SetActivityMessage struct {
	Command string          `json:"cmd"`
	Args    SetActivityArgs `json:"args"`
	Nonce   string          `json:"nonce"`
}

type SetActivityArgs struct {
	PID      int             `json:"pid"`
	Activity ActivityPayload `json:"activity"`
}

type ActivityPayload struct {
	Details    string      `json:"details,omitempty"`
	State      string      `json:"state,omitempty"`
	Timestamps *Timestamps `json:"timestamps,omitempty"`
	Assets     *Assets     `json:"assets,omitempty"`
}

type Timestamps struct {
	Start int64 `json:"start,omitempty"`
}

type Assets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type DiscordRPC struct {
	conn     net.Conn
	clientID string
}

func NewDiscordRPC(clientID string) *DiscordRPC {
	return &DiscordRPC{clientID: clientID}
}

func (d *DiscordRPC) Connect() error {
	for i := 0; i < 10; i++ {
		pipePath := d.getPipePath(i)
		conn, err := net.Dial("unix", pipePath)
		if err == nil {
			d.conn = conn
			return d.handshake()
		} else {
			fmt.Printf("Error: failed to dial unix with pipe path: %v\n", pipePath)
			fmt.Printf("net.Dial error is : %v\n", err)
		}
	}
	return fmt.Errorf("failed to connect to Discord")
}

func (d *DiscordRPC) getPipePath(pipe int) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, pipe)
	default:
		tmpDir := os.Getenv("XDG_RUNTIME_DIR")
		if tmpDir == "" {
			tmpDir = os.Getenv("TMPDIR")
		}
		if tmpDir == "" {
			tmpDir = "/tmp"
		}
		return filepath.Join(tmpDir, fmt.Sprintf("discord-ipc-%d", pipe))
	}
}

func (d *DiscordRPC) handshake() error {
	handshake := HandshakeMessage{
		Version:  "1",
		ClientID: d.clientID,
	}

	return d.sendMessage(OpHandshake, handshake)
}

func (d *DiscordRPC) reconnect() error {
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			// let us see why
			fmt.Printf("Warning: failed to close existing connection: %v\n", err)
		}
		d.conn = nil
	}
	return d.Connect()
}

func (d *DiscordRPC) SetActivity(activity ActivityPayload) error {
	msg := SetActivityMessage{
		Command: "SET_ACTIVITY",
		Args: SetActivityArgs{
			PID:      os.Getpid(),
			Activity: activity,
		},
		Nonce: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	err := d.sendMessage(OpFrame, msg)
	if err != nil {
		// to fix :Error updating rich presence: write unix ->/var/folders/nb/l9h53lq15wd6xtccvkz737_00000gn/T/discord-ipc-0: write: broken pipe
		if reconErr := d.reconnect(); reconErr != nil {
			return fmt.Errorf("failed to update activity and reconnect failed: %v, %v", err, reconErr)
		}
		return d.sendMessage(OpFrame, msg)
	}
	return nil
}

func (d *DiscordRPC) sendMessage(opcode int, payload interface{}) error {
	if d.conn == nil {
		return fmt.Errorf("no connection available")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Discord RPC protocol: opcode (4 bytes) + length (4 bytes) + data
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(opcode))
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(data)))

	if _, err := d.conn.Write(header); err != nil {
		return err
	}

	if _, err := d.conn.Write(data); err != nil {
		return err
	}

	return nil
}

func (d *DiscordRPC) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func getClientID() string {
	if clientID := os.Getenv("DISCORD_CLIENT_ID"); clientID != "" {
		return clientID
	}

	// For CI, this will be done via Github Actions and with ldflags
	if DefaultClientID != "" && DefaultClientID != "PLACEHOLDER_CLIENT_ID" {
		return DefaultClientID
	}

	// For Running Locally
	fmt.Println("No Discord Client ID available!")
	fmt.Println()
	fmt.Println("This binary was not built with a Discord Client ID.")
	fmt.Println("Please either:")
	fmt.Println("Download the official release from GitHub -> https://github.com/siddarthkay/discord-show-local-time/releases ")
	fmt.Println("OR")
	fmt.Println("Set DISCORD_CLIENT_ID environment variable with your own Discord Application ID")
	fmt.Println()
	fmt.Println("To create your own Discord Application:")
	fmt.Println("• Go to: https://discord.com/developers/applications")
	fmt.Println("• Click 'New Application' and give it a name")
	fmt.Println("• Copy the 'Application ID' from the General Information page")
	fmt.Println("• Set it as: export DISCORD_CLIENT_ID='your_application_id'")
	os.Exit(1)
	return ""
}

func validateClientID(clientID string) bool {
	// Client IDs are typically 18-19 digit numbers
	if len(clientID) < 17 || len(clientID) > 20 {
		return false
	}

	for _, char := range clientID {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func main() {
	fmt.Println("Discord Time Rich Presence")
	fmt.Println("=============================")

	clientID := getClientID()

	if !validateClientID(clientID) {
		fmt.Printf("Warning: '%s' doesn't look like a valid Discord Client ID\n", clientID)
		fmt.Println("Discord Client IDs are typically 18-19 digit numbers")
		fmt.Print("Continue anyway? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("Exiting...")
			os.Exit(1)
		}
	}

	fmt.Println("Connecting to Discord...")

	rpc := NewDiscordRPC(clientID)
	if err := rpc.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		fmt.Println()
		fmt.Println("Troubleshooting:")
		fmt.Println("• Make sure Discord Desktop app is running (not browser)")
		fmt.Println("• Enable 'Display current activity as a status message' in Discord Settings")
		fmt.Println("• Try restarting Discord completely")
		os.Exit(1)
	}
	defer func(rpc *DiscordRPC) {
		err := rpc.Close()
		if err != nil {
			fmt.Printf("Failed to close discord client: %v\n", err)
		}
	}(rpc)

	fmt.Println("Connected to Discord!")
	fmt.Println("Starting time updates (Ctrl+C to stop)...")
	fmt.Println()

	updatePresence(rpc)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			updatePresence(rpc)
		}
	}
}

func updatePresence(rpc *DiscordRPC) {
	now := time.Now()
	timeStr := now.Format("3:04 PM MST")
	dateStr := now.Format("Mon, Jan 2")

	activity := ActivityPayload{
		Details: fmt.Sprintf("It's %s", timeStr),
		State:   fmt.Sprintf("🗓️ %s", dateStr),
		Timestamps: &Timestamps{
			Start: now.Unix(),
		},
		Assets: &Assets{
			LargeImage: "clock_icon",
			LargeText:  "Local Time Display",
			SmallImage: "time_small",
			SmallText:  "Live Time",
		},
	}

	if err := rpc.SetActivity(activity); err != nil {
		fmt.Printf("Error updating rich presence: %v\n", err)
	} else {
		fmt.Printf("Updated presence: %s on %s\n", timeStr, dateStr)
	}
}
