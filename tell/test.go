package tell

import (
    "log"
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)

func sendTest(id int) (err error) {
    config.Log.Debug("AgentID", "=", config.AgentID)
    //defer close(waitForTell)
    defer waitForTell.Done()

	botToken := yaml.Agents.Agent.BotToken
	chatID := yaml.Agents.Agent.BotChatID
    config.Log.Debug("BotToken", "=", botToken)
    config.Log.Debug("BotChatID", "=", chatID)

/* allowed format
https://core.telegram.org/api/entities

messageEntityBold => <b>bold</b>, <strong>bold</strong>, **bold**
messageEntityItalic => <i>italic</i>, <em>italic</em> *italic*
messageEntityCode » => <code>code</code>, `code`
messageEntityStrike => <s>strike</s>, <strike>strike</strike>, <del>strike</del>, ~~strike~~
messageEntityUnderline => <u>underline</u>
messageEntityPre » => <pre language="c++">code</pre>,
*/

	message := `This is a test message from Prox
<i>this should be in italic</i>
<b>this should be in bold</b>
<code>this should be in code</code>
<pre language="json">{
      "type": "tun",
      "tag": "tun-in",
      "interface_name": "tun4",
      "mtu": 9000,
      "stack": "mixed",
      "sniff": true,
      "inet4_address": "192.168.3.2/30",
      "endpoint_independent_nat": true
}</pre>
`

	err = sendMsg(message)
    if err != nil {
		config.Log.Error("failed to send the message", "error",  err)
        return fmt.Errorf("sendTest() / sendTelegramMessage()failed  %w", err)
	} else {
		config.Log.Info("the message sent successfully", "id", id)
		log.Printf("the message sent successfully")
	}
    return
}

