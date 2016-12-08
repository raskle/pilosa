package bench

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
)

// called in BagentCommand:Run
func WriteResultToDb(result map[string]interface{}, bm Benchmark) int {
	// Write the result of a benchmark run to a postgres database.
	// Also include the benchmark name, current timestamp, benchmark parameters.
	// Returns number of rows written (sanity check).
	result_json, err := json.Marshal(result)
	bm_json, err := json.Marshal(bm)

	// TODO: dont open new connection for each result
	db, err := sql.Open("postgres", "postgres://abernstein@localhost/pilosabench?sslmode=disable")
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer db.Close()

	rows, err := db.Query("WITH rows AS (INSERT INTO results (name, datetime, params, data) VALUES ($1, CURRENT_TIMESTAMP, $2, $3) RETURNING 1) SELECT count(*) FROM rows;",
		fmt.Sprintf("%T", bm), bm_json, result_json)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			fmt.Println(err)
			return 0
		}
	}

	return count
}
