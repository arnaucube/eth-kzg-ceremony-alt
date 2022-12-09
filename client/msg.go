package client

import "fmt"

type errorMsg struct {
	Message string `json:"message"`
}

type MsgStatus struct {
	LobbySize        uint64 `json:"lobby_size"`
	NumContributions uint64 `json:"num_contributions"`
	SequencerAddress string `json:"sequencer_address"`
}

func (m *MsgStatus) String() string {
	return fmt.Sprintf("Sequencer status:\n  Lobby size: %d\n  NumContributions: %d\n  SequencerAddress: %s\n",
		m.LobbySize, m.NumContributions, m.SequencerAddress)
}

type MsgRequestLink struct {
	EthAuthURL    string `json:"eth_auth_url"`
	GithubAuthURL string `json:"github_auth_url"`
}

type IDToken struct {
	Exp      uint64 `json:"exp"`
	Nickname string `json:"nickname"`
	Provider string `json:"provider"`
	Sub      string `json:"sub"`
}

type MsgAuthCallback struct {
	IDToken   IDToken `json:"id_token"`
	SessionID string  `json:"session_id"`
}

type MsgContributeReceipt struct {
	Receipt   string `json:"receipt"`
	Signature string `json:"signature"`
}

func (m MsgContributeReceipt) String() string {
	return fmt.Sprintf("Contribute Receipt:\n  Receipt: %s\n  Signature: %s\n",
		m.Receipt, m.Signature)
}
