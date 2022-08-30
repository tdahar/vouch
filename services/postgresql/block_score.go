package postgresql

/*

This file together with the model, has all the needed methods to interact with the epoch_metrics table of the database

*/

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

var (
	CREATE_SCORE_TABLE = `
		CREATE TABLE IF NOT EXISTS t_score_metrics(
			f_slot INT,
			f_label VARCHAR(100),
			f_score FLOAT,
			f_duration FLOAT,
			CONSTRAINT PK_SlotAddr PRIMARY KEY (f_slot,f_label));`

	InsertNewScore = `
		INSERT INTO t_score_metrics (	
			f_slot, 
			f_label, 
			f_score,
			f_duration)
		VALUES ($1, $2, $3, $4);`
)

// in case the table did not exist
func (p *PostgresDBService) createScoreMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, CREATE_SCORE_TABLE)
	if err != nil {
		return errors.Wrap(err, "error creating score metrics table")
	}
	return nil
}

func (p *PostgresDBService) InsertNewScore(slot int, label string, score float64, duration float64) error {

	_, err := p.psqlPool.Exec(p.ctx, InsertNewScore, slot, label, score, duration)

	if err != nil {
		return errors.Wrap(err, "error inserting row in score metrics table")
	}
	return nil
}
