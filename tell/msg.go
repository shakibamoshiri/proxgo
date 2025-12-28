package tell

import (
    "fmt"
    "time"
    "encoding/json"
    "net/http"
    "bytes"
    "context"
    "io"

    "github.com/shakibamoshiri/proxgo/config"
)

/* allowed format
https://core.telegram.org/api/entities

messageEntityBold => <b>bold</b>, <strong>bold</strong>, **bold**
messageEntityItalic => <i>italic</i>, <em>italic</em> *italic*
messageEntityCode » => <code>code</code>, `code`
messageEntityStrike => <s>strike</s>, <strike>strike</strike>, <del>strike</del>, ~~strike~~
messageEntityUnderline => <u>underline</u>
messageEntityPre » => <pre language="c++">code</pre>,
*/


func sendMsg(message string) error {
    agents, _ := yaml.Agents.Load()
    botToken := agents.Agent.BotToken
    chatID := agents.Agent.BotChatID

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]any{
		"chat_id": chatID,
		"text":    message,
        "parse_mode": "HTML",
        "disable_notification": false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sendMsg() %w", err)
	}
    config.Log.Info("sendTelegramMessage", "jsonData", payload["chat_id"].(int64))

    client := &http.Client{
        // Timeout: (time.Second * config.ClientTimeout),
        Timeout: (time.Second * 3),
    }

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("sendMsg() %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %s", resp.Status)
	}

	return nil
}

func sendMsgContext(ctx context.Context, message string) error {
    agents, _ := yaml.Agents.Load()
    botToken := agents.Agent.BotToken
    chatID := agents.Agent.BotChatID

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]any{
		"chat_id": chatID,
		"text":    message,
        "parse_mode": "HTML",
        "disable_notification": false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sendMsg() %w", err)
	}
    config.Log.Info("payload", "chat_id", payload["chat_id"].(int64))

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    const subject = "Send http POST request with context "
    config.Log.Info(subject + "...")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        config.Log.Error(subject + "failed / http.DefaultClient.Do(req)")
        config.Log.Debug(subject + "failed", "http.DefaultClient.Do(req)", err)
        return err
    }
    defer resp.Body.Close()

    var body bytes.Buffer
    io.Copy(&body, resp.Body)
    if resp.StatusCode != http.StatusOK {
        config.Log.Error("Telegram error", "response status code", resp.StatusCode)
        config.Log.Debug("Telegram error", "response body", body.String())
        return fmt.Errorf("telegram error %d", resp.StatusCode)
    }

    config.Log.Debug(subject + "... done")
	return nil
}
