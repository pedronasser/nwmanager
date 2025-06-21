# The initial version
if [ ! -f .env ]
then
  export $(cat .env | xargs)
fi

# My favorite from the comments. Thanks @richarddewit & others!
set -a && source .env && set +a

flyctl secrets set --app=$APP_NAME DISCORD_BOT_TOKEN=$DISCORD_BOT_TOKEN
flyctl secrets set --app=$APP_NAME DISCORD_APP_ID=$DISCORD_APP_ID
flyctl secrets set --app=$APP_NAME DISCORD_GUILD_ID=$DISCORD_GUILD_ID
flyctl secrets set --app=$APP_NAME MODULES=$MODULES
flyctl secrets set --app=$APP_NAME MONGO_URI=$MONGO_URI
