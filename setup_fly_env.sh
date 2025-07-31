# The initial version
if [ ! -f .env ]
then
  export $(cat .env | xargs)
fi

set -a && source ./.env && set +a

flyctl secrets set --app=$APP_NAME DISCORD_BOT_TOKEN=$DISCORD_BOT_TOKEN
flyctl secrets set --app=$APP_NAME DISCORD_APP_ID=$DISCORD_APP_ID
flyctl secrets set --app=$APP_NAME DISCORD_GUILD_ID=$DISCORD_GUILD_ID
flyctl secrets set --app=$APP_NAME MODULES=$MODULES
flyctl secrets set --app=$APP_NAME MONGO_URI=$MONGO_URI
flyctl secrets set --app=$APP_NAME GUILD_NAME=$GUILD_NAME
flyctl secrets set --app=$APP_NAME DB_PREFIX=$DB_PREFIX
flyctl secrets set --app=$APP_NAME ADMIN_ROLE_ID=$ADMIN_ROLE_ID