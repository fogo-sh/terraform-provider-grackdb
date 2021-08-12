package types

type DiscordBot struct {
	ID      string          `json:"id"`
	Account *DiscordAccount `json:"account"`
}

type DiscordAccount struct {
	ID            string      `json:"id"`
	DiscordID     string      `json:"discordId"`
	Username      string      `json:"username"`
	Discriminator string      `json:"discriminator"`
	Owner         *User       `json:"owner"`
	Bot           *DiscordBot `json:"bot"`
}
