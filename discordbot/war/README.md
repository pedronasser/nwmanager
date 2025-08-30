# Módulo War - Sistema de Gerenciamento de Guerras

## Visão Geral

O módulo War implementa um sistema completo de gerenciamento de guerras para a guild, permitindo que administradores criem guerras e os jogadores respondam sobre sua participação.

## Funcionalidades Implementadas

### ✅ Fase 1: Estrutura Base e Configuração
- ✅ Módulo `war` criado e registrado
- ✅ Configuração padrão implementada
- ✅ Estrutura War atualizada no banco de dados
- ✅ Sistema de módulos integrado

### ✅ Fase 2: Comando de Criação de Guerra
- ✅ Slash command `/criar-guerra` implementado
- ✅ Modal com validações para:
  - Nome do Forte
  - Tipo de Guerra (Ataque/Defesa)
  - Guild Oponente
  - Data (DD/MM/AAAA)
  - Horário (HH:MM)
- ✅ Validações de permissões (apenas admins)
- ✅ Validações de formato de data/hora
- ✅ Persistência no banco de dados

### ✅ Fase 3: Sistema de Publicação
- ✅ Mensagem no canal de guerras com:
  - Embed rico com informações da guerra
  - Contadores dinâmicos de participação
  - Botão de edição para administradores
- ✅ Sistema de DMs para jogadores via tickets:
  - Embed informativo
  - Botões de resposta (Sim/Não/Talvez/Banco)
- ✅ Contadores em tempo real

### ✅ Fase 4: Sistema de Respostas dos Jogadores
- ✅ Botões de participação funcionais:
  - ✅ "Sim" (verde)
  - ❌ "Não" (vermelho)
  - 🤔 "Talvez" (cinza)
  - 🪑 "Banco" (azul)
- ✅ Atualização automática de:
  - Banco de dados
  - Contadores na mensagem do canal
  - Mensagem privada do jogador
- ✅ Sistema de mudança de resposta

### ✅ Fase 5: Sistema de Cleanup Automático
- ✅ Rotina de verificação (executa a cada minuto)
- ✅ Arquivamento automático quando horário da guerra chega
- ✅ Cleanup de mensagens:
  - Remoção da mensagem do canal
  - Remoção das mensagens dos tickets
- ✅ Logs de atividade

## Arquitetura

```
discordbot/war/
├── module.go      # Configuração e inicialização do módulo
├── handlers.go    # Manipuladores de interações Discord
├── helpers.go     # Funções auxiliares para embeds e mensagens
├── routines.go    # Rotina de cleanup automático
├── constants.go   # Constantes e configurações
└── types.go       # Tipos específicos do módulo
```

## Configuração

### Variáveis de Ambiente

```bash
# Adicionar "war" à lista de módulos
MODULES=globals,register,events,voice_channel,admin_commands,ticket,war

# Canal onde as mensagens de guerra são postadas
WAR_CHANNEL_ID=123456789012345678
```

### Configuração no Banco de Dados

```json
{
  "war": {
    "enabled": true,
    "war_channel_id": "123456789012345678"
  }
}
```

## Como Usar

### Para Administradores

1. **Criar Guerra**: Use o comando `/criar-guerra`
2. **Preencher Formulário**: Complete os campos do modal
3. **Confirmar**: A guerra será criada e publicada automaticamente

### Para Jogadores

1. **Receber Notificação**: DM será enviada via ticket
2. **Responder**: Clique nos botões para indicar participação
3. **Alterar Resposta**: Pode mudar a resposta quantas vezes necessário

## Banco de Dados

### Coleção: `wars`

```javascript
{
  _id: ObjectId,
  fort_name: "Nome do Forte",
  war_type: "attack" | "defense",
  opponent_guild: "Guild Oponente",
  scheduled_at: Date,
  created_at: Date,
  created_by: "discord_user_id",
  archived_at: Date, // quando arquivada
  
  // Participações dos jogadores
  participations: {
    "player_discord_id": "yes" | "no" | "maybe" | "bench"
  },
  
  // IDs das mensagens para cleanup
  channel_message_id: "message_id",
  player_messages: {
    "player_discord_id": "message_id"
  },
  
  status: "active" | "archived"
}
```

## Comandos Implementados

### `/criar-guerra`
- **Descrição**: Cria uma nova guerra para a guild
- **Permissão**: Apenas administradores
- **Parâmetros**: Modal com campos para nome, tipo, oponente, data e hora

## Botões e Interações

### Botões de Participação
- `war_participate:yes:WAR_ID` - Confirma participação
- `war_participate:no:WAR_ID` - Recusa participação
- `war_participate:maybe:WAR_ID` - Talvez participe
- `war_participate:bench:WAR_ID` - Disponível como reserva

### Botões de Administração
- `war_edit:WAR_ID` - Editar guerra (em desenvolvimento)

## Logs e Monitoramento

### Logs Importantes
- Criação de guerras
- Respostas de participação
- Mudanças de resposta
- Arquivamento automático
- Erros de envio de mensagens

### Exemplo de Logs
```
War cleanup routine started with interval: 1m0s
Successfully archived war: Forte do Vento (Forte do Vento vs Inimigos Unidos)
Error sending war DM to player João: message channel not found
```

## Tratamento de Erros

### Cenários Tratados
- ❌ Usuário sem permissão
- ❌ Formato de data/hora inválido
- ❌ Data no passado
- ❌ Guerra não encontrada
- ❌ Guerra já arquivada
- ❌ Jogador sem ticket
- ❌ Falhas de rede/Discord

### Feedback ao Usuário
- Mensagens efêmeras para erros
- Confirmações de sucesso
- Status visual nos embeds

## Próximas Funcionalidades (Roadmap)

### 🚧 Em Desenvolvimento
- [ ] Edição de guerras
- [ ] Cancelamento de guerras
- [ ] Relatórios de participação
- [ ] Notificações antes da guerra

### 🔮 Futuro
- [ ] Guerras recorrentes
- [ ] Templates de guerra
- [ ] Dashboard web
- [ ] Integração com calendário
- [ ] Estatísticas históricas

## Troubleshooting

### Problemas Comuns

1. **Guerra não aparece no canal**
   - Verificar se `WAR_CHANNEL_ID` está configurado
   - Verificar permissões do bot no canal

2. **Jogadores não recebem DM**
   - Verificar se jogadores têm ticket channel
   - Verificar se estão marcados como ativos

3. **Cleanup não funciona**
   - Verificar logs da rotina de cleanup
   - Verificar se horário da guerra passou

### Comandos de Debug

```bash
# Verificar módulo ativo
grep -r "War module is enabled" logs/

# Verificar guerras ativas
mongo nwmanager --eval "db.guild_wars.find({status: 'active'})"

# Verificar cleanup
tail -f logs/ | grep "cleanup"
```

---

**Status**: ✅ **Implementado e Funcional**  
**Versão**: 1.0  
**Data**: 21/08/2025
