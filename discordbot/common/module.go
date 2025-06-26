package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"nwmanager/database"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrModuleAlreadyExists = errors.New("module already exists")
)

type ModuleContext struct {
	Context context.Context
	session *discordgo.Session
	db      database.Database
	configs map[string]any
}

func (ctx *ModuleContext) Session() *discordgo.Session {
	return ctx.session
}

func (ctx *ModuleContext) DB() database.Database {
	return ctx.db
}

func (ctx *ModuleContext) Config(moduleName string) any {
	config, exists := ctx.configs[moduleName]
	if !exists {
		panic(fmt.Sprintf("Module %s config not found", moduleName))
	}
	return config
}

type Module[T any] interface {
	DefaultConfig() T
	Name() string
	Setup(ctx *ModuleContext, config T) (bool, error)
}

func NewModuleManager(guildName string, db database.Database, dg *discordgo.Session) *ModuleManager {
	return &ModuleManager{
		guildName: guildName,
		modules:   make(map[string]Module[any]),
		db:        db,
		session:   dg,
	}
}

type ModuleManager struct {
	guildName string
	modules   map[string]Module[any]
	db        database.Database
	session   *discordgo.Session
}

func (m *ModuleManager) RegisterModule(name string, module Module[any]) error {
	if _, exists := m.modules[name]; exists {
		return ErrModuleAlreadyExists
	}
	m.modules[name] = module
	return nil
}

// loadModulesConfig loads modules config from the modules.json file
// and returns a map of module names to their configurations.
func (m *ModuleManager) loadModulesConfig(ctx context.Context) (configs map[string]any, err error) {
	// Load configs from db
	result := m.db.Collection("config").FindOne(ctx, bson.M{
		"guild": m.guildName,
	})
	if result.Err() != nil {
		log.Printf("Failed to load %s config from database: %v", m.guildName, result.Err())
	}

	var dbConfigs map[string]any
	if err := result.Decode(&dbConfigs); err != nil {
		log.Printf("Error decoding config for guild %s: %v", m.guildName, err)
		return nil, fmt.Errorf("error decoding config for guild %s: %w", m.guildName, err)
	}

	configs = make(map[string]any)
	for name, module := range m.modules {
		config := module.DefaultConfig()
		configs[name] = config

		if dbConfig, exists := dbConfigs[name]; exists {
			data, err := bson.Marshal(dbConfig)
			if err != nil {
				log.Printf("Error marshalling config for module %s: %v", name, err)
				continue
			}

			if err := bson.Unmarshal(data, &config); err != nil {
				log.Printf("Error unmarshalling config for module %s: %v", name, err)
				continue
			}
		}

		log.Println("Loaded config for module:", name)
	}

	return configs, nil
}

func (m *ModuleManager) Run(ctx context.Context) {
	configs, err := m.loadModulesConfig(ctx)
	if err != nil {
		log.Println("Error loading modules config:", err)
		return
	}

	mctx := &ModuleContext{
		Context: ctx,
		session: m.session,
		db:      m.db,
		configs: configs,
	}

	for name, module := range m.modules {

		configs := configs[name]
		running, err := module.Setup(mctx, configs)
		if err != nil {
			log.Printf("Error running module %s: %v", name, err)
		}
		if running {
			log.Printf("Module \"%s\" started", name)
		}
	}
}
