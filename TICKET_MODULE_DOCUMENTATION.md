# Ticket Module Documentation

## Overview

The `ticket` module is a Discord bot component designed to automatically create and manage personal ticket channels for guild members. It integrates with the existing registration system and provides an interface for players to interact with the guild administration through dedicated ticket channels.

## Architecture Analysis

### Based on existing modules structure:
- **Events module**: Manages event creation, participation, and lifecycle with embedded messages and button interactions
- **Register module**: Handles member registration flow with step-by-step processes and role management
- **Common patterns**: Both modules use handlers, configuration, and database integration following the established patterns

## Key Features

### 1. Automatic Ticket Creation
- Monitors guild members for `MEMBER_ROLE_ID` role
- Creates ticket channels automatically when members receive the role
- Uses format: `${playerClassEmoji} ${playerIGN}` for channel naming
- Integrates with existing player data from the register module

### 2. Ticket Management Interface
- Single embedded message per ticket with interactive buttons
- Persistent UI that handles all ticket-related operations
- Admin and player-specific command access control

### 3. Member Role Monitoring
- Background routine to check member role status
- Automatic ticket deletion when `MEMBER_ROLE_ID` is removed
- Automatic ticket recreation when role is re-assigned

## Technical Implementation Plan

### Module Structure
Following the established pattern from events/register modules:

```
discordbot/
└── ticket/
    ├── module.go           # Main module implementation
    ├── handlers.go         # Interaction handlers for buttons/modals
    ├── helpers.go          # Utility functions
    ├── routines.go         # Background monitoring routines
    ├── constants.go        # Constants and emoji definitions
    └── setup.go           # Channel setup and initialization
```

### Database Integration

#### New Collection: `tickets`
```go
type Ticket struct {
    ID            primitive.ObjectID `bson:"_id" json:"id"`
    DiscordID     string             `bson:"discord_id" json:"discord_id"`
    ChannelID     string             `bson:"channel_id" json:"channel_id"`
    MessageID     string             `bson:"message_id" json:"message_id"`
    PlayerIGN     string             `bson:"player_ign" json:"player_ign"`
    PlayerClass   string             `bson:"player_class" json:"player_class"`
    CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
    LastUpdatedAt time.Time          `bson:"last_updated_at" json:"last_updated_at"`
    IsActive      bool               `bson:"is_active" json:"is_active"`
}
```

### Module Configuration
```go
type TicketConfig struct {
    Enabled             bool   `json:"enabled"`
    TicketCategoryID    string `json:"ticket_category_id"`
    CheckInterval       int    `json:"check_interval_seconds"`
    
    // Class emoji mappings
    ClassEmojis         map[string]string `json:"class_emojis"`
    
    // Thread settings
    BuildThreadTimeout  int    `json:"build_thread_timeout_hours"`
    QuestionThreadTimeout int  `json:"question_thread_timeout_hours"`
}

// Note: MemberRoleID and AdminRoleID are accessed from globals config:
// globalCfg, _ := ctx.Config("globals").(*globals.GlobalsConfig)
// memberRoleID := globalCfg.MemberRoleID
// adminRoleID := globalCfg.AdminRoleID
```

### Core Handlers Implementation

#### Button Handlers Map
Following the pattern from events/register modules:

```go
var handlers = map[string]func(ctx *common.ModuleContext, i *discordgo.InteractionCreate){
    "ticket:send_build":       handleSendBuild,
    "ticket:send_question":    handleSendQuestion,
    "ticket:change_class":     handleChangeClass,
    "ticket:view_build":       handleViewBuild,
    "ticket:close_thread":     handleCloseThread,
    "ticket:submit_build":     handleSubmitBuild,
    "select:class_selection":  handleClassSelection,
}
```

### Command Button Implementations

#### 1. Enviar Build Button
```go
func handleSendBuild(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
    // Get player data from database
    player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
    if err != nil || player == nil {
        discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
        return
    }
    
    // Create thread for build submission
    threadName := fmt.Sprintf("build-%s", player.IGN)
    thread, err := ctx.Session().MessageThreadStart(i.ChannelID, i.Message.ID, threadName, 1440) // 24 hours
    if err != nil {
        discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao criar thread para build.", 5*time.Second)
        return
    }
    
    // Send instructions and confirm button in thread
    setupBuildThread(ctx, thread.ID, player.IGN)
    discordutils.ReplyEphemeralMessage(ctx.Session(), i, fmt.Sprintf("Thread criada: <#%s>", thread.ID), 5*time.Second)
}
```

#### 2. Enviar Dúvida Button
```go
func handleSendQuestion(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
    player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
    if err != nil || player == nil {
        discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
        return
    }
    
    threadName := fmt.Sprintf("duvida-%s", player.IGN)
    thread, err := ctx.Session().MessageThreadStart(i.ChannelID, i.Message.ID, threadName, 1440)
    if err != nil {
        discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao criar thread para dúvida.", 5*time.Second)
        return
    }
    
    setupQuestionThread(ctx, thread.ID, player.IGN)
    discordutils.ReplyEphemeralMessage(ctx.Session(), i, fmt.Sprintf("Thread criada: <#%s>", thread.ID), 5*time.Second)
}
```

#### 3. Trocar Classe Guerra Button
```go
func handleChangeClass(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
    config := GetModuleConfig(ctx)
    globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)
    
    // Check if user is admin or the ticket owner
    isAdmin := hasRole(i.Member, globalConfig.AdminRoleID)
    isOwner := isTicketOwner(ctx, i.ChannelID, i.Member.User.ID)
    
    if !isAdmin && !isOwner {
        discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não tem permissão para usar este comando.", 5*time.Second)
        return
    }
    
    // Create class selection dropdown
    classOptions := createClassSelectOptions(config.ClassEmojis)
    
    discordutils.SendInteractiveMessage(ctx.Session(), i, "select:class_selection", "Selecione a nova classe:",
        discordgo.ActionsRow{
            Components: []discordgo.MessageComponent{
                discordgo.SelectMenu{
                    CustomID:    "select:class_selection",
                    MenuType:    discordgo.StringSelectMenu,
                    Placeholder: "Escolha uma classe...",
                    MinValues:   &[]int{1}[0],
                    MaxValues:   1,
                    Options:     classOptions,
                },
            },
        },
    )
}
```

#### 4. Ver Build Atual Button
```go
func handleViewBuild(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
    player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
    if err != nil || player == nil {
        discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
        return
    }
    
    // Get latest build from player's build thread or database
    buildInfo := getCurrentBuildInfo(ctx, player)
    
    embed := &discordgo.MessageEmbed{
        Title:       fmt.Sprintf("Build Atual - %s", player.IGN),
        Description: buildInfo,
        Color:       0x00ff00,
        Timestamp:   time.Now().Format(time.RFC3339),
    }
    
    err = ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Embeds: []*discordgo.MessageEmbed{embed},
            Flags:  discordgo.MessageFlagsEphemeral,
        },
    })
    if err != nil {
        log.Printf("Error responding to view build: %v", err)
    }
}
```

### Background Monitoring Routine

#### Member Role Monitoring
```go
func memberRoleMonitoringRoutine(ctx *common.ModuleContext) {
    config := GetModuleConfig(ctx)
    globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)
    
    ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            log.Println("Checking member roles for ticket management...")
            
            // Get all guild members
            members, err := ctx.Session().GuildMembers(globalConfig.GuildID, "", 1000)
            if err != nil {
                log.Printf("Error fetching guild members: %v", err)
                continue
            }
            
            // Get all active tickets
            activeTickets, err := getAllActiveTickets(ctx)
            if err != nil {
                log.Printf("Error fetching active tickets: %v", err)
                continue
            }
            
            // Check each member's role status
            for _, member := range members {
                err := processmemberRoleStatus(ctx, member, activeTickets, globalConfig.MemberRoleID)
                if err != nil {
                    log.Printf("Error processing member %s: %v", member.User.ID, err)
                }
            }
            
        case <-ctx.Context.Done():
            return
        }
    }
}
```

### Integration Points

#### 1. Register Module Integration
- Hook into registration approval process
- Automatically create ticket when member role is assigned
- Transfer player data (IGN, class, etc.) to ticket system

#### 2. Player Type Updates
- Update `Player` type to include ticket channel reference
- Sync ticket data with player database records
- Handle class changes through both systems

#### 3. Discord Permissions
- Ticket channels visible only to player and admins
- Thread creation permissions for build/question features
- Role-based access control for admin functions

### Channel Setup and Management

#### Ticket Channel Creation
```go
func createTicketChannel(ctx *common.ModuleContext, member *discordgo.Member, player *types.Player) (*discordgo.Channel, error) {
    config := GetModuleConfig(ctx)
    globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)
    
    classEmoji := getClassEmoji(player.WarClass[0]) // Assuming primary class
    channelName := fmt.Sprintf("%s %s", classEmoji, player.IGN)
    
    channel, err := ctx.Session().GuildChannelCreateComplex(globalConfig.GuildID, discordgo.GuildChannelCreateData{
        Name:     channelName,
        Type:     discordgo.ChannelTypeGuildText,
        ParentID: config.TicketCategoryID,
        PermissionOverwrites: []*discordgo.PermissionOverwrite{
            {
                ID:   globalConfig.EveryoneRoleID,
                Type: discordgo.PermissionOverwriteTypeRole,
                Deny: discordgo.PermissionViewChannel,
            },
            {
                ID:    member.User.ID,
                Type:  discordgo.PermissionOverwriteTypeMember,
                Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
            },
            {
                ID:    globalConfig.AdminRoleID,
                Type:  discordgo.PermissionOverwriteTypeRole,
                Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionManageMessages,
            },
        },
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to create ticket channel: %w", err)
    }
    
    // Create and send the main ticket message
    err = setupTicketMessage(ctx, channel, player)
    if err != nil {
        // Clean up channel if message setup fails
        ctx.Session().ChannelDelete(channel.ID)
        return nil, fmt.Errorf("failed to setup ticket message: %w", err)
    }
    
    return channel, nil
}
```

### Error Handling and Resilience

#### Database Consistency
- Ensure ticket records are cleaned up when channels are deleted
- Handle database errors gracefully with retry mechanisms
- Maintain data consistency between player and ticket records

#### Discord API Limitations
- Handle rate limiting for channel creation/deletion
- Implement proper error handling for permission issues
- Graceful degradation when Discord features are unavailable

### Testing Strategy

#### Unit Tests
- Test individual handler functions
- Mock Discord interactions for isolated testing
- Validate database operations independently

#### Integration Tests
- Test complete ticket lifecycle (creation, interaction, deletion)
- Verify role monitoring routine functionality
- Test integration with register module workflows

### Performance Considerations

#### Scalability
- Efficient database queries for large member counts
- Pagination for guild member retrieval
- Caching for frequently accessed player data

#### Resource Management
- Proper cleanup of inactive tickets
- Thread timeout management
- Memory efficient handling of Discord interactions

### Configuration and Deployment

#### Environment Variables
```bash
TICKET_CATEGORY_ID="123456789"
TICKET_CHECK_INTERVAL=300
BUILD_THREAD_TIMEOUT=24
QUESTION_THREAD_TIMEOUT=12
```

Note: `MEMBER_ROLE_ID` and `ADMIN_ROLE_ID` are already configured in the globals module.

#### Default Configuration
```go
func (s *TicketModule) DefaultConfig() *TicketConfig {
    var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)
    
    return &TicketConfig{
        Enabled:               IsModuleEnabledFromEnv,
        TicketCategoryID:      os.Getenv("TICKET_CATEGORY_ID"),
        CheckInterval:         300, // 5 minutes default
        BuildThreadTimeout:    24,  // 24 hours
        QuestionThreadTimeout: 12,  // 12 hours
        ClassEmojis: map[string]string{
            "tank":     "🛡️",
            "dps":      "⚔️",
            "heal":     "💚",
            "support":  "🔧",
        },
    }
}

// Note: MemberRoleID and AdminRoleID are accessed from globals:
// globalCfg, _ := ctx.Config("globals").(*globals.GlobalsConfig)
// memberRoleID := globalCfg.MemberRoleID
// adminRoleID := globalCfg.AdminRoleID
```

## Migration and Rollout Plan

### Phase 1: Core Infrastructure
1. Implement basic module structure and configuration
2. Create database collections and types
3. Set up basic channel creation/deletion functionality

### Phase 2: Interactive Features
1. Implement ticket message UI with buttons
2. Add thread creation for builds and questions
3. Implement class change functionality

### Phase 3: Integration and Monitoring
1. Integrate with register module workflow
2. Implement member role monitoring routine
3. Add comprehensive error handling and logging

### Phase 4: Testing and Optimization
1. Comprehensive testing of all features
2. Performance optimization and scalability improvements
3. Documentation and deployment procedures

## Future Enhancements

### Advanced Features
- Build comparison and history tracking
- Automated build validation against game meta
- Integration with external build planning tools
- Advanced analytics and reporting for guild management

### User Experience Improvements
- Rich embed formatting for build displays
- Image optimization and storage for build screenshots
- Automated reminders for incomplete registrations
- Advanced search and filtering for ticket management
