package storage

import "time"

type Connection struct {
	ID              int64     `json:"id"`
	Address         string    `json:"address"`
	Nickname        string    `json:"nickname"`
	LastConnectedAt time.Time `json:"last_connected_at"`
	CreatedAt       time.Time `json:"created_at"`
}

func (s *Store) SaveConnection(c *Connection) error {
	if c.ID > 0 {
		_, err := s.db.Exec(`UPDATE connections SET address = ?, nickname = ?, last_connected_at = ? WHERE id = ?`,
			c.Address, c.Nickname, c.LastConnectedAt.Format(time.RFC3339), c.ID)
		return err
	}
	_, err := s.db.Exec(`INSERT INTO connections (address, nickname, last_connected_at, created_at) VALUES (?, ?, ?, ?)`,
		c.Address, c.Nickname, c.LastConnectedAt.Format(time.RFC3339), c.CreatedAt.Format(time.RFC3339))
	return err
}

func (s *Store) ListConnections() ([]*Connection, error) {
	rows, err := s.db.Query(`SELECT id, address, nickname, last_connected_at, created_at FROM connections ORDER BY last_connected_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []*Connection
	for rows.Next() {
		var c Connection
		var lastConnected, created string
		if err := rows.Scan(&c.ID, &c.Address, &c.Nickname, &lastConnected, &created); err != nil {
			return nil, err
		}
		c.LastConnectedAt, _ = time.Parse(time.RFC3339, lastConnected)
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		conns = append(conns, &c)
	}
	return conns, rows.Err()
}

func (s *Store) GetConnection(id int64) (*Connection, error) {
	var c Connection
	var lastConnected, created string
	err := s.db.QueryRow(`SELECT id, address, nickname, last_connected_at, created_at FROM connections WHERE id = ?`, id).
		Scan(&c.ID, &c.Address, &c.Nickname, &lastConnected, &created)
	if err != nil {
		return nil, err
	}
	c.LastConnectedAt, _ = time.Parse(time.RFC3339, lastConnected)
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &c, nil
}

func (s *Store) DeleteConnection(id int64) error {
	_, err := s.db.Exec(`DELETE FROM connections WHERE id = ?`, id)
	return err
}
