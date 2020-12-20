# tgServerBridge
Telegram - Server Bridge for up-down files, exec commands easily

# Setup
* Open <a href="https://t.me/botfather" target="_blank">@botfather</a> on Telegram > Create bot, take note for bot token

* Send message any json bot on Telegram for find user id (Example <a href="https://t.me/JsonDumpBot" target="_blank">@JsonDumpBot</a>) > take note for from-id

      export token="<yourBotToken>"
      export adminid="<yourUserID>"
   
   
Run
   
    ./bot
 
With screen: (ctrl + a + d on background )

    screen -S myBot ./bot
    
    
# Commands

/help > Show help message

/ls > list dir

/pwd > where am i

/cd > change dir

/exec <command> ... > exec command | realtime |

\<anyText\> if text avaible and file; upload for user || if text avaible and dir; cd dir else show error

\<anyDoc\> upload for server
  


