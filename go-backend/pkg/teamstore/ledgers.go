package teamstore

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

var (
	ErrLedgerAlreadyLinked = errors.New("ledger already linked to team")
	ErrLedgerNotLinked       = errors.New("ledger not linked to team")
	ErrLastLedger            = errors.New("team must keep at least one ledger")
	ErrUnauthorized          = errors.New("unauthorized")
)

// TeamLedger is a ledger associated with a team (F36).
type TeamLedger struct {
	LedgerID  string    `json:"ledgerId"`
	AddedAt   time.Time `json:"addedAt"`
	AddedByID string    `json:"addedById,omitempty"`
}

func (s *Store) listLedgerIDs(teamID string) ([]string, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return nil, ErrTeamNotFound
	}
	rows, err := s.db.Query(
		`SELECT ledger_id FROM team_ledgers WHERE team_id = ? ORDER BY added_at ASC`, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var lid string
		if err := rows.Scan(&lid); err != nil {
			return nil, err
		}
		out = append(out, lid)
	}
	return out, rows.Err()
}

func (s *Store) listLedgers(teamID string) ([]TeamLedger, error) {
	tid, err := strconv.ParseUint(teamID, 10, 64)
	if err != nil {
		return nil, ErrTeamNotFound
	}
	rows, err := s.db.Query(
		`SELECT ledger_id, added_at, COALESCE(added_by_id, 0) FROM team_ledgers WHERE team_id = ? ORDER BY added_at ASC`,
		tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TeamLedger
	for rows.Next() {
		var tl TeamLedger
		var addedBy uint64
		if err := rows.Scan(&tl.LedgerID, &tl.AddedAt, &addedBy); err != nil {
			return nil, err
		}
		if addedBy > 0 {
			tl.AddedByID = strconv.FormatUint(addedBy, 10)
		}
		out = append(out, tl)
	}
	return out, rows.Err()
}

func (s *Store) insertLedgers(tx *sql.Tx, teamID uint64, creatorID string, ledgerIDs []string) error {
	cid, _ := strconv.ParseUint(creatorID, 10, 64)
	for _, lid := range ledgerIDs {
		if lid == "" {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO team_ledgers (team_id, ledger_id, added_by_id) VALUES (?, ?, ?)`,
			teamID, lid, cid,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) syncPrimaryLedger(tx *sql.Tx, teamID uint64, ledgerIDs []string) error {
	primary := ""
	if len(ledgerIDs) > 0 {
		primary = ledgerIDs[0]
	}
	_, err := tx.Exec(`UPDATE teams SET ledger_id = ? WHERE id = ?`, primary, teamID)
	return err
}

// AddLedger links a multi-person ledger to a team (creator only).
func (s *Store) AddLedger(teamID, ledgerID, actorID string) (*Team, error) {
	if ledgerID == "" {
		return nil, ErrInvalidTeam
	}
	team, err := s.GetByID(teamID)
	if err != nil {
		return nil, err
	}
	if team.CreatorID != actorID {
		return nil, ErrUnauthorized
	}
	ids, err := s.listLedgerIDs(teamID)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		if id == ledgerID {
			return nil, ErrLedgerAlreadyLinked
		}
	}
	tid, _ := strconv.ParseUint(teamID, 10, 64)
	cid, _ := strconv.ParseUint(actorID, 10, 64)
	if _, err := s.db.Exec(
		`INSERT INTO team_ledgers (team_id, ledger_id, added_by_id) VALUES (?, ?, ?)`,
		tid, ledgerID, cid,
	); err != nil {
		return nil, err
	}
	if team.LedgerID == "" {
		_, _ = s.db.Exec(`UPDATE teams SET ledger_id = ? WHERE id = ?`, ledgerID, tid)
	}
	return s.GetByID(teamID)
}

// RemoveLedger unlinks a ledger (creator only; at least one must remain).
func (s *Store) RemoveLedger(teamID, ledgerID, actorID string) (*Team, error) {
	team, err := s.GetByID(teamID)
	if err != nil {
		return nil, err
	}
	if team.CreatorID != actorID {
		return nil, ErrUnauthorized
	}
	ids, err := s.listLedgerIDs(teamID)
	if err != nil {
		return nil, err
	}
	if len(ids) <= 1 {
		return nil, ErrLastLedger
	}
	found := false
	for _, id := range ids {
		if id == ledgerID {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrLedgerNotLinked
	}
	tid, _ := strconv.ParseUint(teamID, 10, 64)
	if _, err := s.db.Exec(`DELETE FROM team_ledgers WHERE team_id = ? AND ledger_id = ?`, tid, ledgerID); err != nil {
		return nil, err
	}
	remaining, _ := s.listLedgerIDs(teamID)
	primary := ""
	if len(remaining) > 0 {
		primary = remaining[0]
	}
	_, _ = s.db.Exec(`UPDATE teams SET ledger_id = ? WHERE id = ?`, primary, tid)
	return s.GetByID(teamID)
}

func (s *Store) attachLedgers(t *Team) {
	ledgers, err := s.listLedgers(t.ID)
	if err != nil || len(ledgers) == 0 {
		if t.LedgerID != "" {
			t.LedgerIDs = []string{t.LedgerID}
			t.Ledgers = []TeamLedger{{LedgerID: t.LedgerID}}
		}
		return
	}
	t.Ledgers = ledgers
	t.LedgerIDs = make([]string, len(ledgers))
	for i, l := range ledgers {
		t.LedgerIDs[i] = l.LedgerID
	}
	if len(t.LedgerIDs) > 0 {
		t.LedgerID = t.LedgerIDs[0]
	}
}
