package hierarchical

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/phuangpheth/hierarchical/database"
)

var ErrUnknownSyllabus = errors.New("unknown syllabus")

type Syllabus struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Term      string     `json:"term"`
	Relations []Relation `json:"parents"`
}

type Relation struct {
	ParentID string `json:"parentId"`
	ChildID  string `json:"childId"`
}

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetByID(ctx context.Context, id string) (*Syllabus, error) {
	syllabus, err := getSyllabusByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	relations, err := getRelationByChildID(ctx, s.db, syllabus.ID)
	if err != nil {
		return nil, err
	}
	syllabus.Relations = relations
	return syllabus, nil
}

func getSyllabusByID(ctx context.Context, db *database.DB, id string) (*Syllabus, error) {
	query, args, err := sq.Select("id", "name", "term").
		From("syllabuses").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(ctx, query, args...)
	s, err := scanSyllabus(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUnknownSyllabus
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func getRelationByChildID(ctx context.Context, db *database.DB, parentID string) ([]Relation, error) {
	query, args, err := sq.Select("child_id", "parent_id").
		From("syllabus_relations").
		Where(sq.Eq{"parent_id": parentID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rs := make([]Relation, 0)
	collection := func(rows *sql.Rows) error {
		p, err := scanRelation(rows.Scan)
		if err != nil {
			return err
		}
		rs = append(rs, p)
		return nil
	}
	return rs, db.RunQuery(ctx, query, collection, args...)
}

func scanRelation(scan func(...any) error) (p Relation, _ error) {
	return p, scan(&p.ChildID, &p.ParentID)
}

func scanSyllabus(scan func(...any) error) (s Syllabus, _ error) {
	return s, scan(&s.ID, &s.Name, &s.Term)
}
