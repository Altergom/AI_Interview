package weixin

import "sync"

// Account 一个已登录的微信 bot 账号。
type Account struct {
	AccountID string // ilink_bot_id
	Token     string // bot_token
	BaseURL   string // 登录确认返回的 baseurl，可空（用默认）
	UserID    string // 扫码用户的 ilink_user_id
}

// Store token 存储。本期为内存版，accountID → Account。
type Store struct {
	mu       sync.RWMutex
	accounts map[string]*Account
}

func NewStore() *Store {
	return &Store{accounts: make(map[string]*Account)}
}

// Save 落库一个账号（登录确认后调用）。
func (s *Store) Save(a *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[a.AccountID] = a
}

// Get 按 accountID 取账号。
func (s *Store) Get(accountID string) (*Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accounts[accountID]
	return a, ok
}

// List 返回全部账号快照。
func (s *Store) List() []*Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		out = append(out, a)
	}
	return out
}

// Tokens 返回全部账号 token，供 get_bot_qrcode 的 local_token_list 去重用。
func (s *Store) Tokens() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.accounts))
	for _, a := range s.accounts {
		if a.Token != "" {
			out = append(out, a.Token)
		}
	}
	return out
}
