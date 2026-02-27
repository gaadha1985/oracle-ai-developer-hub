package oracle

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jasperan/picooraclaw/pkg/config"
)

func inspectOverview(db *sql.DB, agentID string) {
	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Println("  PicoOraClaw Oracle AI Database Inspector")
	fmt.Println("=============================================================")
	fmt.Println()

	tables := []struct {
		name  string
		label string
	}{
		{"PICO_MEMORIES", "Memories"},
		{"PICO_SESSIONS", "Sessions"},
		{"PICO_TRANSCRIPTS", "Transcripts"},
		{"PICO_STATE", "State"},
		{"PICO_DAILY_NOTES", "Daily Notes"},
		{"PICO_PROMPTS", "Prompts"},
		{"PICO_CONFIG", "Config"},
		{"PICO_META", "Meta"},
	}

	fmt.Println("  Table                  Rows")
	fmt.Println("  ─────────────────────  ────")

	totalRows := 0
	for _, t := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", t.name)).Scan(&count)
		if err != nil {
			fmt.Printf("  %-23s  error: %v\n", t.label, err)
			continue
		}
		totalRows += count
		bar := strings.Repeat("█", min(count, 40))
		if count > 0 {
			fmt.Printf("  %-23s %4d  %s\n", t.label, count, bar)
		} else {
			fmt.Printf("  %-23s %4d  (empty)\n", t.label, count)
		}
	}

	fmt.Println("  ─────────────────────  ────")
	fmt.Printf("  %-23s %4d\n", "Total", totalRows)

	// Show latest memories
	fmt.Println()
	fmt.Println("  Recent Memories (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows, err := db.Query(`
		SELECT memory_id, content, importance, category,
		       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created
		FROM PICO_MEMORIES
		WHERE agent_id = :1
		ORDER BY created_at DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows.Close()
		hasRows := false
		for rows.Next() {
			hasRows = true
			var memID, created string
			var content, category sql.NullString
			var importance float64
			if err := rows.Scan(&memID, &content, &importance, &category, &created); err != nil {
				continue
			}
			text := "(empty)"
			if content.Valid {
				text = content.String
			}
			cat := ""
			if category.Valid && category.String != "" {
				cat = fmt.Sprintf(" [%s]", category.String)
			}
			fmt.Printf("  %s  %.1f%s  %s\n", created, importance, cat, text)
		}
		if !hasRows {
			fmt.Println("  (no memories stored yet)")
		}
	}

	// Show latest transcript entries
	fmt.Println()
	fmt.Println("  Recent Transcripts (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows2, err := db.Query(`
		SELECT role, content,
		       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created
		FROM PICO_TRANSCRIPTS
		WHERE agent_id = :1
		ORDER BY id DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows2.Close()
		hasRows := false
		for rows2.Next() {
			hasRows = true
			var role, created string
			var content sql.NullString
			if err := rows2.Scan(&role, &content, &created); err != nil {
				continue
			}
			text := "(empty)"
			if content.Valid {
				text = content.String
			}
			fmt.Printf("  %s  %-10s  %s\n", created, role, text)
		}
		if !hasRows {
			fmt.Println("  (no transcripts yet)")
		}
	}

	// Show latest sessions
	fmt.Println()
	fmt.Println("  Recent Sessions (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows3, err := db.Query(`
		SELECT session_key, summary,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_SESSIONS
		WHERE agent_id = :1
		ORDER BY updated_at DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows3.Close()
		hasRows := false
		for rows3.Next() {
			hasRows = true
			var sessionKey, updated string
			var summary sql.NullString
			if err := rows3.Scan(&sessionKey, &summary, &updated); err != nil {
				continue
			}
			sumText := "(no summary)"
			if summary.Valid && summary.String != "" {
				sumText = summary.String
			}
			fmt.Printf("  %s  %-30s  %s\n", updated, sessionKey, sumText)
		}
		if !hasRows {
			fmt.Println("  (no sessions yet)")
		}
	}

	// Show state entries (last 5 by updated_at)
	fmt.Println()
	fmt.Println("  Recent State Entries (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows4, err := db.Query(`
		SELECT state_key, state_value,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_STATE
		WHERE agent_id = :1
		ORDER BY updated_at DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows4.Close()
		hasRows := false
		for rows4.Next() {
			hasRows = true
			var key, updated string
			var value sql.NullString
			if err := rows4.Scan(&key, &value, &updated); err != nil {
				continue
			}
			val := "(null)"
			if value.Valid {
				val = value.String
			}
			fmt.Printf("  %s  %-25s = %s\n", updated, key, val)
		}
		if !hasRows {
			fmt.Println("  (no state entries yet)")
		}
	}

	// Show latest daily notes
	fmt.Println()
	fmt.Println("  Recent Daily Notes (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows5, err := db.Query(`
		SELECT TO_CHAR(note_date, 'YYYY-MM-DD') AS note_day, content,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_DAILY_NOTES
		WHERE agent_id = :1
		ORDER BY note_date DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows5.Close()
		hasRows := false
		for rows5.Next() {
			hasRows = true
			var noteDay, updated string
			var content sql.NullString
			if err := rows5.Scan(&noteDay, &content, &updated); err != nil {
				continue
			}
			text := "(empty)"
			if content.Valid {
				text = content.String
			}
			fmt.Printf("  %s  (updated %s)  %s\n", noteDay, updated, text)
		}
		if !hasRows {
			fmt.Println("  (no daily notes yet)")
		}
	}

	// Show prompts
	fmt.Println()
	fmt.Println("  System Prompts (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows6, err := db.Query(`
		SELECT prompt_name, DBMS_LOB.GETLENGTH(content) AS content_len,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_PROMPTS
		WHERE agent_id = :1
		ORDER BY updated_at DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows6.Close()
		hasRows := false
		for rows6.Next() {
			hasRows = true
			var name, updated string
			var contentLen sql.NullInt64
			if err := rows6.Scan(&name, &contentLen, &updated); err != nil {
				continue
			}
			size := int64(0)
			if contentLen.Valid {
				size = contentLen.Int64
			}
			fmt.Printf("  %s  %-25s  %d chars\n", updated, name, size)
		}
		if !hasRows {
			fmt.Println("  (no prompts stored yet)")
		}
	}

	// Show config entries
	fmt.Println()
	fmt.Println("  Config Entries (last 5):")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows7, err := db.Query(`
		SELECT config_key, config_value,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_CONFIG
		WHERE agent_id = :1
		ORDER BY updated_at DESC
		FETCH FIRST 5 ROWS ONLY`, agentID)
	if err == nil {
		defer rows7.Close()
		hasRows := false
		for rows7.Next() {
			hasRows = true
			var key, updated string
			var value sql.NullString
			if err := rows7.Scan(&key, &value, &updated); err != nil {
				continue
			}
			val := "(null)"
			if value.Valid {
				val = value.String
			}
			fmt.Printf("  %s  %-25s = %s\n", updated, key, val)
		}
		if !hasRows {
			fmt.Println("  (no config entries stored yet)")
		}
	}

	// Show meta
	fmt.Println()
	fmt.Println("  Schema Metadata:")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows8, err := db.Query(`
		SELECT meta_key, meta_value,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_META
		ORDER BY meta_key
		FETCH FIRST 5 ROWS ONLY`)
	if err == nil {
		defer rows8.Close()
		hasRows := false
		for rows8.Next() {
			hasRows = true
			var key, updated string
			var value sql.NullString
			if err := rows8.Scan(&key, &value, &updated); err != nil {
				continue
			}
			val := "(null)"
			if value.Valid {
				val = value.String
			}
			fmt.Printf("  %s  %-30s = %s\n", updated, key, val)
		}
		if !hasRows {
			fmt.Println("  (no metadata yet)")
		}
	}

	fmt.Println()
	fmt.Println("  Tip: Run 'picooraclaw oracle-inspect <table>' for details")
	fmt.Println("       Run 'picooraclaw oracle-inspect memories -s \"query\"' for semantic search")
	fmt.Println()
}

// inspectMemories shows detailed memory entries.
func inspectMemories(db *sql.DB, agentID string, limit int, searchQuery string, oracleCfg *config.OracleDBConfig) {
	fmt.Println()
	if searchQuery != "" {
		fmt.Printf("  Semantic Search: \"%s\"\n", searchQuery)
		fmt.Println("  ─────────────────────────────────────────────────────────")

		sqlQuery := fmt.Sprintf(`
			SELECT memory_id, content, importance, category,
			       ROUND(1 - VECTOR_DISTANCE(embedding,
			         VECTOR_EMBEDDING(%s USING :1 AS DATA), COSINE), 3) AS similarity,
			       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created,
			       access_count
			FROM PICO_MEMORIES
			WHERE agent_id = :2 AND embedding IS NOT NULL
			ORDER BY VECTOR_DISTANCE(embedding,
			  VECTOR_EMBEDDING(%s USING :3 AS DATA), COSINE) ASC
			FETCH FIRST :4 ROWS ONLY`,
			oracleCfg.ONNXModel, oracleCfg.ONNXModel)

		rows, err := db.Query(sqlQuery, searchQuery, agentID, searchQuery, limit)
		if err != nil {
			fmt.Printf("  Search error: %v\n", err)
			return
		}
		defer rows.Close()

		hasRows := false
		for rows.Next() {
			hasRows = true
			var memID, created string
			var content, category sql.NullString
			var importance, similarity float64
			var accessCount int
			if err := rows.Scan(&memID, &content, &importance, &category, &similarity, &created, &accessCount); err != nil {
				continue
			}
			text := "(empty)"
			if content.Valid {
				text = content.String
			}
			cat := ""
			if category.Valid && category.String != "" {
				cat = category.String
			}
			pct := similarity * 100
			fmt.Printf("\n  [%5.1f%% match]  ID: %s\n", pct, memID)
			fmt.Printf("  Created: %s  Importance: %.1f  Category: %s  Accessed: %dx\n",
				created, importance, cat, accessCount)
			fmt.Printf("  Content: %s\n", text)
		}
		if !hasRows {
			fmt.Println("  No matching memories found.")
		}
	} else {
		fmt.Println("  All Memories")
		fmt.Println("  ─────────────────────────────────────────────────────────")

		rows, err := db.Query(`
			SELECT memory_id, content, importance, category,
			       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created,
			       access_count,
			       CASE WHEN embedding IS NOT NULL THEN 'yes' ELSE 'no' END AS has_vec
			FROM PICO_MEMORIES
			WHERE agent_id = :1
			ORDER BY created_at DESC
			FETCH FIRST :2 ROWS ONLY`, agentID, limit)
		if err != nil {
			fmt.Printf("  Query error: %v\n", err)
			return
		}
		defer rows.Close()

		hasRows := false
		for rows.Next() {
			hasRows = true
			var memID, created, hasVec string
			var content, category sql.NullString
			var importance float64
			var accessCount int
			if err := rows.Scan(&memID, &content, &importance, &category, &created, &accessCount, &hasVec); err != nil {
				continue
			}
			text := "(empty)"
			if content.Valid {
				text = content.String
			}
			cat := ""
			if category.Valid && category.String != "" {
				cat = category.String
			}
			fmt.Printf("\n  ID: %s  Vector: %s\n", memID, hasVec)
			fmt.Printf("  Created: %s  Importance: %.1f  Category: %s  Accessed: %dx\n",
				created, importance, cat, accessCount)
			fmt.Printf("  Content: %s\n", text)
		}
		if !hasRows {
			fmt.Println("  No memories stored yet.")
		}
	}
	fmt.Println()
}

// inspectSessions shows stored chat sessions.
func inspectSessions(db *sql.DB, agentID string, limit int) {
	fmt.Println()
	fmt.Println("  Chat Sessions")
	fmt.Println("  ─────────────────────────────────────────────────────────")

	rows, err := db.Query(`
		SELECT session_key, summary,
		       DBMS_LOB.GETLENGTH(messages) AS msg_len,
		       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_SESSIONS
		WHERE agent_id = :1
		ORDER BY updated_at DESC
		FETCH FIRST :2 ROWS ONLY`, agentID, limit)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var sessionKey, created, updated string
		var summary sql.NullString
		var msgLen sql.NullInt64
		if err := rows.Scan(&sessionKey, &summary, &msgLen, &created, &updated); err != nil {
			continue
		}
		size := int64(0)
		if msgLen.Valid {
			size = msgLen.Int64
		}
		fmt.Printf("\n  Session: %s\n", sessionKey)
		fmt.Printf("  Created: %s  Updated: %s  Messages size: %d bytes\n", created, updated, size)
		if summary.Valid && summary.String != "" {
			fmt.Printf("  Summary: %s\n", summary.String)
		}
	}
	if !hasRows {
		fmt.Println("  No sessions stored yet.")
	}
	fmt.Println()
}

// inspectTranscripts shows conversation transcript entries.
func inspectTranscripts(db *sql.DB, agentID string, limit int) {
	fmt.Println()
	fmt.Println("  Conversation Transcripts")
	fmt.Println("  ─────────────────────────────────────────────────────────")

	rows, err := db.Query(`
		SELECT session_key, sequence_num, role, content,
		       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') AS created
		FROM PICO_TRANSCRIPTS
		WHERE agent_id = :1
		ORDER BY id DESC
		FETCH FIRST :2 ROWS ONLY`, agentID, limit)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var role, created string
		var sessionKey sql.NullString
		var seqNum sql.NullInt64
		var content sql.NullString
		if err := rows.Scan(&sessionKey, &seqNum, &role, &content, &created); err != nil {
			continue
		}
		sess := ""
		if sessionKey.Valid {
			sess = sessionKey.String
		}
		text := "(empty)"
		if content.Valid {
			text = content.String
		}
		seq := int64(0)
		if seqNum.Valid {
			seq = seqNum.Int64
		}
		fmt.Printf("  %s  #%d  %-10s  [%s]  %s\n", created, seq, role, sess, text)
	}
	if !hasRows {
		fmt.Println("  No transcripts yet.")
	}
	fmt.Println()
}

// inspectState shows key-value state entries.
func inspectState(db *sql.DB, agentID string) {
	fmt.Println()
	fmt.Println("  Agent State (Key-Value)")
	fmt.Println("  ─────────────────────────────────────────────────────────")

	rows, err := db.Query(`
		SELECT state_key, state_value,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_STATE
		WHERE agent_id = :1
		ORDER BY state_key`, agentID)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var key, updated string
		var value sql.NullString
		if err := rows.Scan(&key, &value, &updated); err != nil {
			continue
		}
		val := "(null)"
		if value.Valid {
			val = value.String
		}
		fmt.Printf("  %-30s = %-40s  (%s)\n", key, val, updated)
	}
	if !hasRows {
		fmt.Println("  No state entries yet.")
	}
	fmt.Println()
}

// inspectDailyNotes shows daily note entries.
func inspectDailyNotes(db *sql.DB, agentID string, limit int) {
	fmt.Println()
	fmt.Println("  Daily Notes")
	fmt.Println("  ─────────────────────────────────────────────────────────")

	rows, err := db.Query(`
		SELECT note_id, TO_CHAR(note_date, 'YYYY-MM-DD') AS note_day, content,
		       CASE WHEN embedding IS NOT NULL THEN 'yes' ELSE 'no' END AS has_vec,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_DAILY_NOTES
		WHERE agent_id = :1
		ORDER BY note_date DESC
		FETCH FIRST :2 ROWS ONLY`, agentID, limit)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var noteID, noteDay, hasVec, updated string
		var content sql.NullString
		if err := rows.Scan(&noteID, &noteDay, &content, &hasVec, &updated); err != nil {
			continue
		}
		text := "(empty)"
		if content.Valid {
			text = content.String
		}
		fmt.Printf("\n  Date: %s  ID: %s  Vector: %s  Updated: %s\n", noteDay, noteID, hasVec, updated)
		fmt.Printf("  Content: %s\n", text)
	}
	if !hasRows {
		fmt.Println("  No daily notes yet.")
	}
	fmt.Println()
}

// inspectPrompts shows system prompts stored in Oracle.
// If nameFilter is non-empty, prints full content of that prompt only.
// If nameFilter is empty, lists all prompts with their sizes.
func inspectPrompts(db *sql.DB, agentID string, nameFilter ...string) {
	fmt.Println()

	// Single prompt: show full content.
	if len(nameFilter) > 0 && nameFilter[0] != "" {
		name := nameFilter[0]
		fmt.Printf("  Prompt: %s\n", name)
		fmt.Println("  ─────────────────────────────────────────────────────────")
		var content sql.NullString
		var updated string
		err := db.QueryRow(`
			SELECT content, TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
			FROM PICO_PROMPTS
			WHERE agent_id = :1 AND UPPER(prompt_name) = UPPER(:2)`,
			agentID, name).Scan(&content, &updated)
		if err != nil {
			fmt.Printf("  Not found: %v\n", err)
			fmt.Println()
			return
		}
		fmt.Printf("  Updated: %s\n\n", updated)
		if content.Valid {
			fmt.Println(content.String)
		} else {
			fmt.Println("  (empty)")
		}
		fmt.Println()
		return
	}

	// List view: show all prompts with size.
	fmt.Println("  System Prompts")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	fmt.Println("  (use: oracle-inspect prompts <name>  to view full content)")
	fmt.Println()

	rows, err := db.Query(`
		SELECT prompt_name, DBMS_LOB.GETLENGTH(content) AS content_len,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_PROMPTS
		WHERE agent_id = :1
		ORDER BY prompt_name`, agentID)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var name, updated string
		var contentLen sql.NullInt64
		if err := rows.Scan(&name, &contentLen, &updated); err != nil {
			continue
		}
		size := int64(0)
		if contentLen.Valid {
			size = contentLen.Int64
		}
		fmt.Printf("  %-25s  %5d chars  (%s)\n", name, size, updated)
	}
	if !hasRows {
		fmt.Println("  No prompts stored yet.")
	}
	fmt.Println()
}

// inspectConfig shows runtime config stored in Oracle.
func inspectConfig(db *sql.DB, agentID string) {
	fmt.Println()
	fmt.Println("  Stored Config")
	fmt.Println("  ─────────────────────────────────────────────────────────")

	rows, err := db.Query(`
		SELECT config_key, config_value,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_CONFIG
		WHERE agent_id = :1
		ORDER BY config_key`, agentID)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var key, updated string
		var value sql.NullString
		if err := rows.Scan(&key, &value, &updated); err != nil {
			continue
		}
		val := "(null)"
		if value.Valid {
			val = value.String
		}
		fmt.Printf("  %-30s = %-40s  (%s)\n", key, val, updated)
	}
	if !hasRows {
		fmt.Println("  No config entries stored yet.")
	}
	fmt.Println()
}

// inspectMeta shows schema metadata.
func inspectMeta(db *sql.DB) {
	fmt.Println()
	fmt.Println("  Schema Metadata")
	fmt.Println("  ─────────────────────────────────────────────────────────")

	rows, err := db.Query(`
		SELECT meta_key, meta_value,
		       TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI') AS updated
		FROM PICO_META
		ORDER BY meta_key`)
	if err != nil {
		fmt.Printf("  Query error: %v\n", err)
		return
	}
	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		var key, updated string
		var value sql.NullString
		if err := rows.Scan(&key, &value, &updated); err != nil {
			continue
		}
		val := "(null)"
		if value.Valid {
			val = value.String
		}
		fmt.Printf("  %-30s = %-30s  (%s)\n", key, val, updated)
	}
	if !hasRows {
		fmt.Println("  No metadata entries yet.")
	}

	// Also show ONNX model info
	fmt.Println()
	fmt.Println("  ONNX Models")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows2, err := db.Query(`SELECT model_name, mining_function, algorithm FROM user_mining_models ORDER BY model_name`)
	if err == nil {
		defer rows2.Close()
		hasModels := false
		for rows2.Next() {
			hasModels = true
			var modelName, miningFunc, algo string
			if err := rows2.Scan(&modelName, &miningFunc, &algo); err != nil {
				continue
			}
			fmt.Printf("  %-25s  %-15s  %s\n", modelName, miningFunc, algo)
		}
		if !hasModels {
			fmt.Println("  No ONNX models loaded.")
		}
	}

	// Show vector indexes
	fmt.Println()
	fmt.Println("  Vector Indexes")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	rows3, err := db.Query(`SELECT index_name, table_name FROM user_indexes WHERE index_name LIKE 'IDX_PICO_%VEC' ORDER BY index_name`)
	if err == nil {
		defer rows3.Close()
		hasIdx := false
		for rows3.Next() {
			hasIdx = true
			var idxName, tableName string
			if err := rows3.Scan(&idxName, &tableName); err != nil {
				continue
			}
			fmt.Printf("  %-30s  on %s\n", idxName, tableName)
		}
		if !hasIdx {
			fmt.Println("  No vector indexes found.")
		}
	}
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
