package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type User struct {
	ID         int64
	Provider   string
	ProviderID string
	Username   string
}

type Team struct {
	ID         int
	Name, Code string
	FifaRank   int
}

type Store struct {
	db *sql.DB
}

type LeaderboardEntry struct {
	Username string
	Points   int
}

type SideName struct {
	Username string
	Side     string
}

type Fixture struct {
	ID                   int
	HomeName, AwayName   string
	HomeGoals, AwayGoals *int
	HomeId, AwayId       int
	Kickoff              time.Time
	Played               bool
	HomeSide, AwaySide   []SideName
	Status               string
}

func openStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			provider TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			username TEXT NOT NULL,
			created_at  DATETIME NOT NULL,
			last_login  DATETIME NOT NULL,
			UNIQUE (provider, provider_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT PRIMARY KEY,
			user_id    INTEGER NOT NULL REFERENCES users(id),
			created_at DATETIME NOT NULL,
			expires_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)`,
		`CREATE TABLE IF NOT EXISTS teams (
			id        INTEGER PRIMARY KEY,
			name      TEXT NOT NULL,
			code      TEXT NOT NULL,
			fifa_rank INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS rankings (
			user_id INTEGER NOT NULL REFERENCES users(id),
			team_id INTEGER NOT NULL REFERENCES teams(id),
			rank INTEGER NOT NULL,
			PRIMARY KEY (user_id, team_id),
			UNIQUE (user_id, rank)
		)`,
		`CREATE TABLE IF NOT EXISTS FIXTURES (
			id INTEGER PRIMARY KEY,
			home_team_id INTEGER NOT NULL REFERENCES teams(id),
			away_team_id INTEGER NOT NULL REFERENCES teams(id),
			kickoff DATETIME NOT NULL,
			status TEXT NOT NULL,
			home_goals INTEGER,
			away_goals INTEGER
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	var n int
	s.db.QueryRow(`SELECT COUNT(*) FROM teams`).Scan(&n)
	if n == 0 {
		teams := []struct {
			id         int
			name, code string
			fifaRank   int
		}{
			{1, "Spain", "ESP", 2},
			{2, "France", "FRA", 1},
			{3, "England", "ENG", 4},
			{4, "Argentina", "ARG", 3},
			{5, "Brazil", "BRA", 6},
			{6, "Portugal", "POR", 5},
			{7, "Germany", "GER", 10},
			{8, "Netherlands", "NED", 7},
			{9, "Norway", "NOR", 31},
			{10, "Belgium", "BEL", 9},
			{11, "Colombia", "COL", 13},
			{12, "Ecuador", "ECU", 23},
			{13, "Croatia", "CRO", 11},
			{14, "Uruguay", "URU", 17},
			{15, "Morocco", "MAR", 8},
			{16, "Turkiye", "TUR", 22},
			{17, "Switzerland", "SUI", 19},
			{18, "Japan", "JPN", 18},
			{19, "Senegal", "SEN", 14},
			{20, "Mexico", "MEX", 15},
			{21, "United States", "USA", 16},
			{22, "Austria", "AUT", 24},
			{23, "Paraguay", "PAR", 40},
			{24, "Sweden", "SWE", 38},
			{25, "Canada", "CAN", 30},
			{26, "Scotland", "SCO", 43},
			{27, "Algeria", "ALG", 28},
			{28, "Cote d'Ivoire", "CIV", 34},
			{29, "South Korea", "KOR", 25},
			{30, "Australia", "AUS", 27},
			{31, "Czechia", "CZE", 41},
			{32, "Iran", "IRN", 21},
			{33, "Egypt", "EGY", 29},
			{34, "Panama", "PAN", 33},
			{35, "Uzbekistan", "UZB", 50},
			{36, "DR Congo", "COD", 46},
			{37, "Jordan", "JOR", 63},
			{38, "Tunisia", "TUN", 44},
			{39, "Bosnia and Herzegovina", "BIH", 65},
			{40, "Iraq", "IRQ", 57},
			{41, "New Zealand", "NZL", 85},
			{42, "Ghana", "GHA", 74},
			{43, "Saudi Arabia", "KSA", 61},
			{44, "Cabo Verde", "CPV", 69},
			{45, "Haiti", "HAI", 83},
			{46, "South Africa", "RSA", 60},
			{47, "Curacao", "CUW", 82},
			{48, "Qatar", "QAT", 55},
		}
		for _, t := range teams {
			if _, err := s.db.Exec(
				`INSERT INTO teams (id, name, code, fifa_rank) VALUES (?, ?, ?, ?)`,
				t.id, t.name, t.code, t.fifaRank); err != nil {
				return err
			}
		}
	}
	var nf int
	s.db.QueryRow(`SELECT COUNT(*) FROM fixtures`).Scan(&nf)
	if nf == 0 {
		now := time.Now().UTC()
		gi := func(n int) *int { return &n } // helper for goal pointers
		fixtures := []struct {
			id, home, away       int
			kickoff              time.Time
			status               string
			homeGoals, awayGoals *int
		}{
			// finished: New Zealand 3-0 Iran
			{1, 41, 32, now.Add(-24 * time.Hour), "FT", gi(3), gi(0)},
			// finished draw: Spain 1-1 England
			{2, 1, 3, now.Add(-12 * time.Hour), "FT", gi(1), gi(1)},
			// upcoming: Switzerland v Mexico
			{3, 17, 20, now.Add(24 * time.Hour), "NS", nil, nil},
		}
		for _, f := range fixtures {
			if _, err := s.db.Exec(
				`INSERT INTO fixtures (id, home_team_id, away_team_id, kickoff, status, home_goals, away_goals)
                         VALUES (?, ?, ?, ?, ?, ?, ?)`,
				f.id, f.home, f.away, f.kickoff, f.status, f.homeGoals, f.awayGoals); err != nil {
				return err
			}
		}
	}
	return nil
}

// store username
func (s *Store) AddUser(provider, providerID string, username string) (*User, error) {
	now := time.Now().UTC()
	_, err := s.db.Exec(`
		INSERT INTO users (provider, provider_id, username, created_at, last_login)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(provider, provider_id) DO UPDATE SET
			username    = excluded.username,
			last_login  = excluded.last_login
	`, provider, providerID, username, now, now)
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRow(
		`SELECT id, provider, provider_id, username FROM users WHERE provider = ? AND provider_id = ?`,
		provider, providerID)
	var u User
	if err := row.Scan(&u.ID, &u.Provider, &u.ProviderID, &u.Username); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) RefreshFixtures(resp *fifaMatchesResp) error {
	codeToId := map[string]int{}
	rows, err := s.db.Query(`SELECT id, code FROM teams`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var code string
		rows.Scan(&id, &code)
		codeToId[code] = id
	}
	rows.Close()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, f := range resp.Results {

		homeTeamId, ok := codeToId[f.Home.IdCountry]
		if !ok {
			continue
		}
		awayTeamId, ok := codeToId[f.Away.IdCountry]
		if !ok {
			continue
		}
		status := "NS"
		if f.MatchStatus == 0 {
			status = "FT"
		}
		var hg, ag *int
		if f.MatchStatus == 0 {
			hg, ag = f.HomeTeamScore, f.AwayTeamScore
		}
		_, err := tx.Exec(
			`INSERT INTO fixtures (id, home_team_id, away_team_id, kickoff, status, home_goals, away_goals)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(id) DO UPDATE SET
                    status     = excluded.status,
                    home_goals = excluded.home_goals,
                    away_goals = excluded.away_goals`,
			f.IdMatch, homeTeamId, awayTeamId, f.Date.UTC(), status, hg, ag)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) FetchFixtures() ([]Fixture, error) {
	var out []Fixture
	rows, err := s.db.Query(`SELECT f.id,f.home_team_id,f.away_team_id, t1.name, t2.name, f.kickoff, f.status, f.home_goals, f.away_goals FROM Fixtures f
		inner join Teams t1 on f.home_team_id = t1.id
		inner join Teams t2 on f.away_team_id = t2.id
		ORDER BY f.kickoff`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f Fixture

		if err := rows.Scan(&f.ID, &f.HomeId, &f.AwayId, &f.HomeName, &f.AwayName, &f.Kickoff, &f.Status, &f.HomeGoals, &f.AwayGoals); err != nil {
			return nil, err
		}

		home, away, err := s.FixtureSides(f.HomeId, f.AwayId)
		if err != nil {
			return nil, err
		}
		for _, name := range home {
			f.HomeSide = append(f.HomeSide, SideName{name, f.HomeName})
		}
		for _, name := range away {
			f.AwaySide = append(f.AwaySide, SideName{name, f.AwayName})
		}
		out = append(out, f)

	}
	// if err := rows.Err(); err != nil {
	// 	rows.Close()
	// 	return nil, err
	// } else {
	// 	rows.Close()
	// }

	// var out []Fixture
	// for _, r := range fixtures {
	// 	f := r.f
	// 	f.Played = r.status == "FT"

	// 	homeColor, awayColor := ""
	// }
	return out, nil
}

func (s *Store) FixtureSides(homeId, awayId int) (home []string, away []string, err error) {
	rows, err := s.db.Query(`SELECT u.username, CASE WHEN rh.rank < ra.rank THEN 'home' ELSE 'away' END
  FROM users u
  JOIN rankings rh ON rh.user_id = u.id AND rh.team_id = ?
  JOIN rankings ra ON ra.user_id = u.id AND ra.team_id = ?
  ORDER BY u.username`, homeId, awayId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var isHome string
		if err := rows.Scan(&name, &isHome); err != nil {
			return nil, nil, err
		}
		if isHome == "home" {
			home = append(home, name)
		} else {
			away = append(away, name)
		}
	}
	return home, away, nil
}

func (s *Store) CreateSession(userID int64, token string, ttl time.Duration) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(
		`INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		token, userID, now, now.Add(ttl))
	return err
}

func (s *Store) UserBySession(token string) (*User, error) {
	row := s.db.QueryRow(`
		SELECT u.id, provider, provider_id, username
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > ?`,
		token, time.Now().UTC())
	var u User
	if err := row.Scan(&u.ID, &u.Provider, &u.ProviderID, &u.Username); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (s *Store) AllTeams() ([]Team, error) {
	rows, err := s.db.Query(`SELECT id, name, code, fifa_rank FROM teams ORDER BY fifa_rank`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.FifaRank); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (s *Store) UserRankings(userID int64) (map[int]int, error) {
	rows, err := s.db.Query(`SELECT team_id, rank FROM rankings WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rankings := make(map[int]int)
	for rows.Next() {
		var teamID int
		var rank int
		if err := rows.Scan(&teamID, &rank); err != nil {
			return nil, err
		}
		rankings[teamID] = rank
	}
	return rankings, nil
}

// save user's rankings (only works if no rankings exist)
func (s *Store) SaveRankings(userID int64, orderedTeams []int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, teamID := range orderedTeams {
		_, err := tx.Exec(`INSERT INTO rankings (user_id, team_id, rank) VALUES (?, ?, ?)`, userID, teamID, i+1)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Leaderboard() ([]LeaderboardEntry, error) {
	rows, err := s.db.Query(
		`SELECT u.username,
    COALESCE((
      SELECT SUM(CASE
        WHEN f.home_goals = f.away_goals THEN 1
        WHEN rh.rank < ra.rank AND f.home_goals > f.away_goals THEN 3
        WHEN ra.rank < rh.rank AND f.away_goals > f.home_goals THEN 3
        ELSE 0
      END)
      FROM fixtures f
      JOIN rankings rh ON rh.user_id = u.id AND rh.team_id = f.home_team_id
      JOIN rankings ra ON ra.user_id = u.id AND ra.team_id = f.away_team_id
      WHERE f.status = 'FT'
    ), 0) AS points
  FROM users u
  WHERE EXISTS (SELECT 1 FROM rankings r WHERE r.user_id = u.id)
  ORDER BY points DESC, u.username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.Points); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
