UPDATE todos 
SET scheduled_start = ?, scheduled_end = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?
