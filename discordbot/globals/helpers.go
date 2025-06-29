package globals

import (
	"nwmanager/discordbot/common"
	"slices"

	"github.com/bwmarrin/discordgo"
)

var knownRoles = map[string]discordgo.Role{}

func GetRoleByName(guild *discordgo.Guild, roleName string) *discordgo.Role {
	if role, ok := knownRoles[roleName]; ok {
		return &role
	}
	for _, role := range guild.Roles {
		if role.Name == roleName {
			knownRoles[roleName] = *role
			return role
		}
	}

	return nil
}

func IsMemberAdmin(ctx *common.ModuleContext, member *discordgo.Member) bool {
	globalCfg := GetModuleConfig(ctx)
	return slices.Contains(member.Roles, globalCfg.AdminRoleID)
}
