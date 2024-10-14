package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
	"xabbo.b7c.io/goearth/shockwave/profile"
	"xabbo.b7c.io/goearth/shockwave/room"
)

// Global variables for dice management, rolling state, mutex, and wait group
var (
	mutex            sync.RWMutex // Used to synchronize access to the map
	MimicIsDisabled  bool
	FollowIsDisabled bool
	ownName          string
	followTarget     string
	mutedDuration    int
	roomID           string
	isMuted          bool
	roomMgr          = room.NewManager(ext)
	profileMgr       = profile.NewManager(ext)
	userFigures      = make(map[string]UserInfo) // Store username -> UserInfo
	users            = map[int]*User{}
	userMap          = make(map[string]struct{ x, y int })
)

type App struct {
	ext    *g.Ext
	assets embed.FS
	log    []string
	logMu  sync.Mutex
	ctx    context.Context
}

type OutfitConfig struct {
	Outfits map[string]OutfitDetails `json:"outfits"`
}

type OutfitDetails struct {
	Figure string `json:"figure"`
}

func (a *App) SaveConfig(config *OutfitConfig) {
	configFilePath := getConfigFilePath()

	file, err := os.Create(configFilePath)
	if err != nil {
		a.logMsg("Error creating outfit save-file: " + err.Error())
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		a.logMsg("Error encoding outfit save-file: " + err.Error())
		return
	}

	a.logMsg("outfit save-file updated successfully")
}

func getConfigFilePath() string {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "Origin-Tools")
	os.MkdirAll(configPath, 0700)
	return filepath.Join(configPath, "outfit_config.json")
}

func (a *App) LoadConfig() *OutfitConfig {
	configFilePath := getConfigFilePath()

	file, err := os.Open(configFilePath)
	if err != nil {
		// Return an empty config with an initialized map
		return &OutfitConfig{
			Outfits: make(map[string]OutfitDetails),
		}
	}
	defer file.Close()

	var config OutfitConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		a.logMsg("Error decoding save-file: " + err.Error())
		// Return an empty config with an initialized map
		return &OutfitConfig{
			Outfits: make(map[string]OutfitDetails),
		}
	}

	// If the config does not contain outfits, ensure the map is initialized
	if config.Outfits == nil {
		config.Outfits = make(map[string]OutfitDetails)
	}

	// Return the loaded config
	return &config
}

// Save a new outfit by name
func (a *App) SaveOutfit(name string, figure string) {
	// Load the existing config
	config := a.LoadConfig()
	if config == nil {
		config = &OutfitConfig{
			Outfits: make(map[string]OutfitDetails),
		}
	}

	// Save the new outfit with the provided name
	config.Outfits[name] = OutfitDetails{
		Figure: figure,
	}

	// Write the updated config to the file
	a.SaveConfig(config)

	// Log success
	a.logMsg("Outfit saved successfully: " + name)
}

func (a *App) SaveOutfitFromUsername(username string, outfitName string) {
	mutex.RLock() // Use read lock when accessing the map
	defer mutex.RUnlock()

	userInfo, found := userFigures[username]
	if found {
		// log.Printf("Found user %s with figure %s and gender %s", username, userInfo.Figure, userInfo.Gender)

		// Apply gender-based formatting for the figure
		var formattedFigure string
		if userInfo.Gender == "m" {
			formattedFigure = "@l@D@Y" + userInfo.Figure + "@E@AM@JH@AH@R@@" // Adjust male formatting
		} else if userInfo.Gender == "f" {
			formattedFigure = "@l@D@Y" + userInfo.Figure + "@E@AF@JH@AH@R@@" // Adjust female formatting
		} else {
			log.Printf("Unknown gender: %s", userInfo.Gender)
			formattedFigure = userInfo.Figure // Fallback for unknown gender
		}

		// Save the outfit with the formatted figure
		a.SaveOutfit(outfitName, formattedFigure)
		log.Printf("Outfit saved for user %s as %s", username, outfitName)
	} else {
		log.Printf("User %s not found", username)
	}
}

// Load all saved outfits
func (a *App) LoadOutfit(name string) *OutfitDetails {
	config := a.LoadConfig()
	if config == nil {
		a.logMsg("No saved outfits found")
		return nil
	}

	outfit, found := config.Outfits[name]
	if !found {
		a.logMsg("Outfit not found: " + name)
		return nil
	}

	// Return the loaded outfit
	return &outfit
}

func (a *App) LoadAndApplyOutfit(name string) {
	// Load the outfit details by name
	outfit := a.LoadOutfit(name)
	if outfit == nil {
		return
	}

	// Apply the saved outfit figure using the same logic as in copyFigureFromUsername
	updateOwnFigure(outfit.Figure)

	a.logMsg("Outfit applied: " + name)
}

func NewApp(ext *g.Ext, assets embed.FS) *App {
	return &App{
		ext:    ext,
		assets: assets,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.setupExt()
	go func() {
		a.runExt()
	}()
}

// CopyOutfit copies the outfit from the selected user
func (a *App) CopyOutfit(username string) error {
	a.copyFigureFromUsername(username) // Call existing logic
	a.logMsg("Copied outfit from " + username)
	return nil
}

// FollowUser sets the follow target to the selected user
func (a *App) FollowUser(username string) error {
	a.followUser(username) // Call existing logic
	a.logMsg("Following " + username)
	return nil
}

// StopFollowingUser sets the follow target to an empty string: "".
func (a *App) StopFollowingUser(username string) error {
	a.unfollowUser(username) // Call existing logic
	a.logMsg("Stopped Following " + username)
	return nil
}

// LoadUsers returns a list of all users in the room
func (a *App) LoadUsers() []string {
	var userList []string
	for _, user := range users {
		userList = append(userList, user.Name) // Extract usernames
	}
	return userList
}

// logMsg adds a message to the log and broadcasts it to the frontend
func (a *App) logMsg(msg string) {
	a.logMu.Lock()
	defer a.logMu.Unlock()
	a.log = append(a.log, msg)

	// Send the log to the frontend
	runtime.EventsEmit(a.ctx, "logUpdate", strings.Join(a.log, "\n"))
}

func (a *App) setupExt() {
	go a.updateRoom()
	a.ext.Intercept(out.CHAT, out.SHOUT, out.WHISPER).With(a.onChatMessage)
	a.ext.Intercept(out.GETFLATINFO).With(a.getRoomId)
	a.ext.Intercept(in.CHAT).With(a.handleChat)
	a.ext.Intercept(in.CHAT_2).With(a.handleWhisper)
	a.ext.Intercept(in.CHAT_3).With(a.handleShout)
	a.ext.Intercept(in.OPC_OK).With(a.handleEnterRoom)
	a.ext.Intercept(in.USERS).With(a.handleUsers)
	a.ext.Intercept(in.LOGOUT).With(a.handleRemoveUser)
	roomMgr.EntitiesAdded(onEntityAdded)
	profileMgr.Updated(onProfileUpdated) // Listen for profile updates
	roomMgr.EntityUpdated(onEntityStatusUpdated)
	a.ext.InterceptAll(func(e *g.Intercept) {
		handleMutePacket(e)
	})
}

func (a *App) updateRoom() {
	for {
		newUsers := make(map[string]struct{ x, y int })
		roomMgr.Entities(func(ent room.Entity) bool {
			newUsers[ent.Name] = struct{ x, y int }{ent.X, ent.Y}
			return true
		})

		userMap = newUsers
		runtime.EventsEmit(a.ctx, "usersUpdated", a.LoadUsers()) // Emit user list

		time.Sleep(time.Second)
	}
}

func (a *App) updateRoomID(roomId string) {
	roomID = roomId
}

func (a *App) getRoomId(e *g.Intercept) {
	// Updates RoomID whenever a new room is loaded
	roomID = string(e.Packet.Data)
	a.updateRoomID(roomID)
	log.Printf("Room ID change detected: %s", roomID)

}

func (a *App) handleEnterRoom(e *g.Intercept) {
	users = make(map[int]*User)
	// Emit event to frontend to update the users list
	runtime.EventsEmit(a.ctx, "usersUpdated", users) // Emit user list
}

func (a *App) handleRemoveUser(e *g.Intercept) {
	s := e.Packet.ReadString()
	index, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	username := getUsername(index)
	if username != "" {
		chatMsg := ChatMessage{
			Timestamp: time.Now(),
			UserID:    index,
			Username:  username,
			Message:   "",
			Type:      "left",
		}
		sendChatMessage(chatMsg)
		delete(users, index)
	}
	// Emit event to frontend to update the users list
	runtime.EventsEmit(a.ctx, "usersUpdated", a.LoadUsers()) // Emit user list
}

func (a *App) handleUsers(e *g.Intercept) {
	for range e.Packet.ReadInt() {
		var user User
		e.Packet.Read(&user)
		if user.Type == 1 {
			users[user.Index] = &user
			chatMsg := ChatMessage{
				Timestamp: time.Now(),
				UserID:    user.Index,
				Username:  user.Name,
				Message:   "",
				Type:      "join",
			}
			sendChatMessage(chatMsg)
		}
	}
	// Emit event to frontend to update the users list
	runtime.EventsEmit(a.ctx, "usersUpdated", a.LoadUsers()) // Emit user list
}

func (a *App) copyFigureFromUsername(username string) {
	copyFigureFromUsername(username) // Call the existing logic to copy the figure
}

func (a *App) followUser(username string) {
	followTarget = username // Call the existing logic to set follow target
}
func (a *App) unfollowUser(username string) {
	followTarget = "" // Call the existing logic to unset follow target
}

func (a *App) getUsers() []User {
	var userList []User
	for _, user := range users {
		userList = append(userList, *user)
	}
	return userList
}

func (a *App) handleChat(e *g.Intercept) {
	handleChatMessage(e, "chat")
}

func (a *App) handleWhisper(e *g.Intercept) {
	handleChatMessage(e, "whisper")
}

func (a *App) handleShout(e *g.Intercept) {
	handleChatMessage(e, "shout")
}

func handleChatMessage(e *g.Intercept, msgType string) {
	index := e.Packet.ReadInt()
	msg := e.Packet.ReadString()
	username := getUsername(index)
	if username == "" {
		username = "Unknown"
	}

	chatMsg := ChatMessage{
		Timestamp: time.Now(),
		UserID:    index,
		Username:  username,
		Message:   msg,
		Type:      msgType,
	}

	sendChatMessage(chatMsg)
}

func (a *App) runExt() {
	defer os.Exit(0)
	a.ext.Run()
}

func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
}

func startMuteTimer(duration int) {
	for duration > 0 {
		log.Printf("Remaining mute time: %d seconds", duration)
		time.Sleep(1 * time.Second) // Sleep for 1 second
		duration--
	}

	// Mute duration finished
	handleMuteEnd()
}

func handleMuteEnd() {
	isMuted = false
}

func handleMutePacket(e *g.Intercept) {
	// Check for the "first muted" packet with header 4069
	if e.Packet.Header.Value == 4069 {
		mutedDuration = e.Packet.ReadInt() // Read the mute duration in seconds
		log.Printf("You are muted for %d seconds.", mutedDuration)
		isMuted = true
		go startMuteTimer(mutedDuration) // Start the mute timer
	}

	// Check for the "trying to chat while muted" packet with header 3285
	if e.Packet.Header.Value == 3285 {
		remainingMuteDuration := e.Packet.ReadInt() // Read the remaining mute duration
		log.Printf("Mute still active, remaining time: %d seconds.", remainingMuteDuration)
	}
}

func (a *App) onChatMessage(e *g.Intercept) {
	msg := e.Packet.ReadString()

	// Process commands based on the message prefix and suffix
	if strings.HasPrefix(msg, ":") {

		command := strings.TrimPrefix(msg, ":")
		parts := strings.Fields(command)

		switch parts[0] {
		case "copy":
			e.Block()
			if len(parts) > 1 {
				username := parts[1]
				// Find the entity and copy the figure
				copyFigureFromUsername(username)
			} else {
				log.Println("No username provided for :copy command.")
			}
		case "follow":
			e.Block()
			if len(parts) > 1 {
				followTarget = parts[1]
				log.Printf("Following user: %s", followTarget)
			} else {
				log.Println("No username provided for :follow command.")
			}
		case "followon":
			e.Block()
			FollowIsDisabled = false
		case "followoff":
			e.Block()
			FollowIsDisabled = true
		case "mimicon":
			e.Block()
			MimicIsDisabled = false
		case "mimicoff":
			e.Block()
			MimicIsDisabled = true
		}
	}
}

func sendChatMessage(msg ChatMessage) {
	if msg.Type == "left" {
		log.Printf("[Leave] - [%s] Left the room", msg.Username)
		return
	}
	if msg.Type == "join" {
		log.Printf("[Join] - [%s] Entered the room", msg.Username)
		return
	}

	formattedMessage := formatMessage(msg.Message)
	log.Printf("[%s] Sent: %s", msg.Username, msg.Message)

	if msg.Username == followTarget {
		if !MimicIsDisabled {
			// sleep random between 1000 and 1500ms
			time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
			log.Printf("[Mimic] Sent Message: %s from follow target: %s", formattedMessage, msg.Username)

			switch msg.Type {
			case "shout":
				ext.Send(out.SHOUT, formattedMessage)
			case "chat":
				ext.Send(out.CHAT, formattedMessage)
			case "whisper":
				ext.Send(out.WHISPER, msg.Username, formattedMessage)
			default:
				// Handle any unexpected message types, if needed
				log.Printf("Unknown message type: %s", msg.Type)
			}
		}
	}
}

func getUsername(index int) string {
	if user, ok := users[index]; ok {
		return user.Name
	}
	return ""
}

// Format the message for MiMiC text
func formatMessage(message string) string {
	var sb strings.Builder
	for i, char := range strings.ToLower(message) {
		if i%2 == 0 {
			sb.WriteRune(unicode.ToUpper(char))
		} else {
			sb.WriteRune(char)
		}
	}

	return sb.String()
}

// Function to handle when an entity's status is updated (i.e., they move)
func onEntityStatusUpdated(e room.EntityUpdateArgs) {
	// Log the new position of the entity
	// log.Printf("Entity %s moved to position X: %d, Y: %d, Z: %.2f", e.Entity.Name, e.Entity.X, e.Entity.Y, e.Entity.Z)

	// If this is the entity being followed, move to their position
	if e.Entity.Name == followTarget {
		if !FollowIsDisabled {
			// log.Printf("Following Target: %s", followTarget)
			moveToPosition(e.Entity.Tile) // Move to the entity's new position
		}
	}
}

// Function to move to a specific tile
func moveToPosition(tile room.Tile) {
	ext.Send(out.MOVE, int16(tile.X), int16(tile.Y))
	// log.Printf("Moving to position X: %d, Y: %d", tile.X, tile.Y)
}

// Search for the entity by username and copy the figure
func copyFigureFromUsername(username string) {
	mutex.RLock() // Use read lock when accessing the map
	defer mutex.RUnlock()

	userInfo, found := userFigures[username]
	if found {
		// log.Printf("Found user %s with figure %s and gender %s", username, userInfo.Figure, userInfo.Gender)

		// Apply gender-based formatting for the figure
		var formattedFigure string
		if userInfo.Gender == "m" {
			formattedFigure = "@l@D@Y" + userInfo.Figure + "@E@AM@JH@AH@R@@" // Adjust male formatting
		} else if userInfo.Gender == "f" {
			formattedFigure = "@l@D@Y" + userInfo.Figure + "@E@AF@JH@AH@R@@" // Adjust female formatting

		} else {
			// log.Printf(userInfo.Gender)
			formattedFigure = userInfo.Figure // Fallback, in case of unknown gender
		}

		updateOwnFigure(formattedFigure) // Apply the formatted figure
	} else {
		log.Printf("User %s not found in the room.", username)
	}
}

// // Handle when entities (users) are added to the room
func onEntityAdded(e room.EntitiesArgs) {
	mutex.Lock() // Lock before modifying the map
	defer mutex.Unlock()

	for _, entity := range e.Entities {
		userFigures[entity.Name] = UserInfo{
			Figure: entity.Figure,
			Gender: entity.Gender, // Assuming entity.Gender contains "M" or "F"
		}
		// log.Printf("[LOG] User %s added with figure %s", entity.Name, entity.Figure)
		// logPlayerPosition(entity.Name, entity.Tile) // Log player position here
	}
}

// Function to log player position
func logPlayerPosition(username string, tile room.Tile) {
	log.Printf("Player %s is at position X: %d, Y: %d, Z: %.2f", username, tile.X, tile.Y, tile.Z)
}

// // Handle when the profile is updated
func onProfileUpdated(args profile.Args) {
	ownName = args.Profile.Name // Store own name for comparison if needed
	log.Printf("Profile updated. Name: %s, Figure: %s", args.Profile.Name, args.Profile.Figure)
	if len(roomID) > 0 {
		ext.Send(out.GOTOFLAT, []byte(roomID))
		log.Printf("Reloaded Room: %s", roomID)
	}
}

// Update your figure using ProfileManager
func updateOwnFigure(newFigure string) {
	ext.Send(out.UPDATE, newFigure)
	// @l@D@Y8410918001285152900925510@JH@AH@R@@
	log.Printf("Updated figure to: %s", newFigure)
}

func (a *App) MimicOn() {
	MimicIsDisabled = false
}

func (a *App) MimicOff() {
	MimicIsDisabled = true
}

func (a *App) FollowOn() {
	FollowIsDisabled = false
}

func (a *App) FollowOff() {
	FollowIsDisabled = true
}

type UserInfo struct {
	Figure string
	Gender string
}
